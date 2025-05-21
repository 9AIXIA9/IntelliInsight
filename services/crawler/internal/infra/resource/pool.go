package resource

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/utils"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	units        chan domain.ResourceUnit //根据状态操作
	strategy     domain.LoadBalancingStrategy
	roundIdx     atomic.Uint32
	stop         chan struct{}
	ips          []string
	dataDirs     []string
	fingerPrints []string
	maxSize      int
	browserPath  string
	headless     bool
	count        int
	countMu      sync.RWMutex
}

func NewResourcePool(conf config.Resource) domain.ResourcePool {
	if len(conf.DataBasePath) == 0 {
		logx.Severef("资源池初始化失败：数据路径为空")
	}

	if conf.MaxPoolSize < 0 {
		conf.MaxPoolSize = 5
	}

	if conf.InitialSize < 0 {
		conf.InitialSize = 2
	}

	pool := &Pool{
		units:        make(chan domain.ResourceUnit, conf.MaxPoolSize),
		strategy:     conf.LoadBalancingStrategy,
		roundIdx:     atomic.Uint32{},
		stop:         make(chan struct{}),
		ips:          conf.IPs,
		dataDirs:     make([]string, conf.MaxPoolSize),
		fingerPrints: make([]string, conf.MaxPoolSize),
		maxSize:      conf.MaxPoolSize,
		browserPath:  conf.BrowserPath,
		headless:     conf.Headless,
		count:        0,
		countMu:      sync.RWMutex{},
	}

	if len(conf.FingerPrints) == 0 {
		pool.fingerPrints = domain.DefaultFingerPrints
	}

	//加载用户数据目录
	pool.loadDataCatalog(conf.DataBasePath)

	pool.initialize(conf.InitialSize)

	pool.AsyncHealthCheck(conf.HealthCheckInterval)

	return pool
}

// 提前加载资源
func (p *Pool) initialize(count int) {
	for i := 0; i < count; i++ {
		unit := p.createUnit()
		if unit != nil {
			p.units <- unit
		}
	}
}

// 创建资源单元
func (p *Pool) createUnit() domain.ResourceUnit {
	p.countMu.Lock()
	if p.count < p.maxSize {
		p.count++
		p.countMu.Unlock()
	} else {
		p.countMu.Unlock()
		return nil
	}

	dataDir, ip, fingerPrint := p.LoadBalancing()
	unit := NewUnit(dataDir, ip, fingerPrint, p.headless, p.browserPath)

	// 对返回值进行检查
	if !unit.CheckHealth() {
		p.countMu.Lock()
		p.count-- // 创建失败，减少计数
		p.countMu.Unlock()
		logx.Error("创建资源单元失败，浏览器初始化异常")
		return nil
	}

	return unit
}

func (p *Pool) GetResource(ctx context.Context) (domain.ResourceUnit, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case unit := <-p.units:
		return p.guaranteedHealth(unit), nil
	default:
		//成功创建
		if unit := p.createUnit(); unit != nil {
			return unit, nil
		}

		//未成功创建则等待获得
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unit := <-p.units:
			return p.guaranteedHealth(unit), nil
		}
	}
}

// 保证unit是健康的
func (p *Pool) guaranteedHealth(unit domain.ResourceUnit) domain.ResourceUnit {
	if unit == nil {
		unit = p.createUnit()
	}

	//健康则直接返回
	if unit.CheckHealth() {
		return unit
	}
	logx.Errorf("获取到不健康的资源，尝试更新")

	// 刷新资源
	return p.RefreshResource(unit)
}

func (p *Pool) ReleaseResource(unit domain.ResourceUnit) {
	if unit == nil {
		return
	}

	threading.GoSafe(func() {
		p.units <- p.guaranteedHealth(unit)
	})
}

// RefreshResource 更新资源单元
func (p *Pool) RefreshResource(unit domain.ResourceUnit) domain.ResourceUnit {
	//为空则创建
	if unit == nil {
		return p.createUnit()
	}

	dataDir, ip, fingerPrint := p.LoadBalancing()
	unit.Refresh(dataDir, ip, fingerPrint)

	// 检查更新后的资源是否健康，但不递归刷新
	if unit.CheckHealth() {
		return unit
	}

	// 如果刷新后仍然不健康，尝试创建新的资源单元
	logx.Errorf("资源刷新后仍不健康，尝试创建新实例")
	if newUnit := NewUnit(dataDir, ip, fingerPrint, p.headless, p.browserPath); newUnit != nil {
		if newUnit.CheckHealth() {
			unit.Close()
			return newUnit
		}

		newUnit.Close()
		return unit
	}

	// 无法创建新实例时返回原实例
	return unit
}

func (p *Pool) Close() {
	close(p.stop)
	close(p.units)
}

// AsyncHealthCheck 启动非阻塞的健康检查
func (p *Pool) AsyncHealthCheck(interval time.Duration) {
	threading.GoSafe(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.CheckUnitsHealth()
			case <-p.stop:
				return
			}
		}
	})
}

// CheckUnitsHealth 检查单元健康
func (p *Pool) CheckUnitsHealth() {
	//抽样检查 保证系统可用性
	for i := 0; i < p.count; i++ {
		select {
		case unit := <-p.units:
			p.units <- p.guaranteedHealth(unit)
		default:
			logx.Infof("目前资源都在使用")
			break
		}
	}
}

// LoadBalancingStrategy 负载均衡策略
func (p *Pool) LoadBalancingStrategy() domain.LoadBalancingStrategy {
	return p.strategy
}

// LoadBalancing 负载均衡策略
func (p *Pool) LoadBalancing() (dataDir string, ip string, fingerPrint string) {
	switch p.strategy {
	case domain.Random:
		return utils.GetRandomly3(p.dataDirs, p.ips, p.fingerPrints)
	case domain.Round:
		return utils.GetByRound3(p.roundIdx.Add(1), p.dataDirs, p.ips, p.fingerPrints)
	default:
		return utils.GetRandomly3(p.dataDirs, p.ips, p.fingerPrints)
	}
}

// 获取该路径下的子文件夹名称 不要扫描太多，最多 2倍的 maxSize
func (p *Pool) loadDataCatalog(basePath string) {
	// 确保基础路径存在
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.MkdirAll(basePath, 0755)
		if err != nil {
			logx.Severef("创建基础路径失败: %v", err)
		}
	}

	// 读取目录内容
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logx.Severef("读取数据目录失败: %v", err)
	}

	// 限制子目录扫描数量
	count := 0
	maxCount := p.maxSize * 2

	// 遍历目录项
	for _, entry := range entries {
		if count >= maxCount {
			break // 达到扫描上限
		}

		if entry.IsDir() {
			// 将子文件夹路径添加到dataPath
			dirPath := filepath.Join(basePath, entry.Name())
			p.dataDirs[count] = dirPath
			count++
		}
	}

	// 如果找到的子目录不足，则创建必要的目录
	for count < p.maxSize {
		dirPath := filepath.Join(basePath, fmt.Sprintf("profile_%d", time.Now().UnixMilli()))
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				logx.Errorf("创建配置目录失败: %v", err)
				count++ // 即使失败也增加计数避免无限循环
				continue
			}
		}
		p.dataDirs[count] = dirPath
		count++
	}
}
