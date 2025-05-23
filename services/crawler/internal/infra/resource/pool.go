package resource

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/control"
	"crawler/internal/domain"
	"crawler/internal/infra/utils"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxRetries = 10
	maxCount   = 20
)

type Pool struct {
	units chan domain.ResourceUnit //根据状态操作

	stop      chan struct{}
	count     int
	countMu   sync.Mutex
	isRunning bool
	runningMu sync.RWMutex

	strategy     domain.LoadBalancingStrategy
	roundIdx     atomic.Uint32
	ips          []string
	dataDirs     []string
	fingerPrints []string
	maxSize      int
	browserPath  string
	headless     bool
}

func MustNewResourcePool(conf config.Resource) domain.ResourcePool {
	if len(conf.DataBasePath) == 0 {
		control.LogSeveref("资源池初始化失败：数据路径为空")
	}

	if conf.MaxPoolSize < 0 {
		conf.MaxPoolSize = 5
	}

	if conf.InitialSize < 0 {
		conf.InitialSize = 2
	}

	pool := &Pool{
		units:        make(chan domain.ResourceUnit, conf.MaxPoolSize),
		stop:         make(chan struct{}),
		count:        0,
		countMu:      sync.Mutex{},
		isRunning:    true,
		runningMu:    sync.RWMutex{},
		strategy:     conf.LoadBalancingStrategy,
		roundIdx:     atomic.Uint32{},
		ips:          conf.IPs,
		dataDirs:     make([]string, conf.MaxPoolSize),
		fingerPrints: make([]string, conf.MaxPoolSize),
		maxSize:      conf.MaxPoolSize,
		browserPath:  conf.BrowserPath,
		headless:     conf.Headless,
	}

	if len(conf.FingerPrints) == 0 {
		pool.fingerPrints = domain.DefaultFingerPrints
	}

	//加载用户数据目录
	if err := pool.loadDataCatalog(conf.DataBasePath); err != nil {
		control.LogSeveref("加载数据目录失败：%v", err)
	}

	if err := pool.initialize(conf.InitialSize); err != nil {
		control.LogSeveref("初始化资源池失败：%v", err)
	}

	pool.AsyncHealthCheck(conf.HealthCheckInterval)

	return pool
}

// 初始化健康的资源单元
func (p *Pool) initialize(count int) error {
	var lastErr error
	n := 0
	for i := 0; i < count+maxRetries; i++ {
		unit, err := p.createHealthyUnit()
		//创建失败则记录错误重试
		if err != nil {
			lastErr = err
			continue
		}

		p.units <- unit
		n++
		if n == count {
			logx.Infof("成功初始化资源池")
			return nil
		}
	}

	//初始化失败
	return fmt.Errorf("初始化失败，重试次数过多：%w", lastErr)
}

// 创建资源单元 -> 只管创建
func (p *Pool) createHealthyUnit() (domain.ResourceUnit, error) {
	p.runningMu.RLock()
	defer p.runningMu.RUnlock()

	if !p.isRunning {
		return nil, errors.New("资源池已关闭")
	}

	//获取健康单元
	p.countMu.Lock()
	if p.count >= p.maxSize {
		p.countMu.Unlock()
		return nil, errors.New("资源池已达最大数目")
	}
	p.count++ // 先增加计数
	p.countMu.Unlock()

	dataDir, ip, fingerPrint := p.loadBalanced()
	unit, err := NewHealthyUnit(dataDir, ip, fingerPrint, p.headless, p.browserPath)
	if err != nil {
		// 创建失败时减少计数
		p.countMu.Lock()
		p.count--
		p.countMu.Unlock()
		return nil, err
	}

	return unit, nil
}

// 销毁 Unit
func (p *Pool) destroyUnit(unit domain.ResourceUnit) {
	p.countMu.Lock()
	defer p.countMu.Unlock()
	p.count--
	if unit == nil {
		return
	}
	unit.Close()
}

// 检查 Unit 健康并尝试刷新，若尝试无果则销毁不健康实例
func (p *Pool) checkAndHandleUnitHealth(unit domain.ResourceUnit) error {
	err := p.checkUnitHealth(unit)
	if err == nil {
		return nil
	}

	//不健康则尝试刷新资源
	var lastErr error

	for i := 0; i < 3; i++ {
		err := p.Refresh(unit)
		if err != nil {
			lastErr = err
			continue
		}

		if err = p.checkUnitHealth(unit); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	//刷新后依旧不健康则放弃该资源 -> 也就是销毁
	logx.Infof("处理不健康资源单位失败，尝试摧毁")
	p.destroyUnit(unit)
	return fmt.Errorf("经过尝试后该资源仍不健康：%w", lastErr)
}

// 检查 Unit 健康
func (p *Pool) checkUnitHealth(unit domain.ResourceUnit) error {
	if unit == nil {
		return errors.New("资源为空")
	}
	return unit.CheckHealth()
}

// AsyncHealthCheck 启动非阻塞的健康检查
func (p *Pool) AsyncHealthCheck(interval time.Duration) {
	threading.GoSafe(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.spotCheckHealth()
			case <-p.stop:
				logx.Infof("已关闭健康检查")
				return
			}
		}
	})
}

// spotCheckHealth 抽样检查单元健康
func (p *Pool) spotCheckHealth() {
	//抽样检查 保证系统可用性
	for i := 0; i < p.count; i++ {
		select {
		case unit := <-p.units:
			if err := p.checkAndHandleUnitHealth(unit); err != nil {
				logx.Errorf("资源不健康，且处理出错：%v", err)
				logx.Infof("放弃该资源")
				p.destroyUnit(unit)
				continue
			}
			p.units <- unit

		default:
			logx.Infof("目前资源都在使用")
		}
	}
}

// Refresh 更新资源单元
func (p *Pool) Refresh(unit domain.ResourceUnit) error {
	//资源为空
	if unit == nil {
		return errors.New("刷新失败，资源为空")
	}

	dataDir, ip, fingerPrint := p.loadBalanced()
	if err := unit.Refresh(dataDir, ip, fingerPrint); err != nil {
		return err
	}

	return nil
}

// Get 获取健康资源
func (p *Pool) Get(ctx context.Context) (domain.ResourceUnit, error) {
	p.runningMu.RLock()
	defer p.runningMu.RUnlock()

	if !p.isRunning {
		return nil, errors.New("资源池已关闭")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case unit := <-p.units:
		logx.Debugf("从资源池中获取资源")
		err := p.checkAndHandleUnitHealth(unit)
		if err != nil {
			return nil, err
		}
		return unit, nil
	default:
		logx.Debugf("在资源池中创建资源")

		//成功创建
		if unit, err := p.createHealthyUnit(); err == nil {
			return unit, nil
		}

		//未成功创建则等待获得
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unit := <-p.units:
			err := p.checkAndHandleUnitHealth(unit)
			if err != nil {
				return nil, err
			}
			return unit, nil
		}
	}
}

// Put 负责释放资源
func (p *Pool) Put(unit domain.ResourceUnit) {
	logx.Debugf("从资源池中获取资源")

	threading.GoSafe(func() {
		//检查资源是否正常
		if err := p.checkAndHandleUnitHealth(unit); err != nil {
			//不正常的资源销毁他
			p.destroyUnit(unit)
			logx.Errorf("释放的资源有误：%v", err)
			return
		}

		//正常的资源放回池中
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		select {
		case p.units <- unit:
			logx.Infof("成功回收资源")
		case <-ctx.Done():
			logx.Infof("回收资源超时，即将摧毁资源")
			p.destroyUnit(unit)
		case <-p.stop:
			logx.Infof("资源池已关闭")
			p.destroyUnit(unit)
		}
	})
}

func (p *Pool) Close(ctx context.Context) {
	p.runningMu.Lock()
	if !p.isRunning {
		p.runningMu.Unlock()
		return // 已经关闭
	}

	p.isRunning = false
	close(p.stop)
	p.runningMu.Unlock()

	logx.Info("正在关闭资源池...")

	// 关闭所有资源，带超时保护
	timeout := time.After(10 * time.Second)
	closed := 0

	for closed < p.count {
		select {
		case unit := <-p.units:
			unit.Close()
			closed++
		case <-timeout:
			logx.Errorf("关闭资源池超时，已关闭 %d/%d 个资源", closed, p.count)
			return
		case <-ctx.Done():
			logx.Errorf("关闭资源池被取消，已关闭 %d/%d 个资源", closed, p.count)
			return
		}
	}

	close(p.units)
	logx.Info("资源池已完全关闭")
}

// LoadBalancingStrategy 负载均衡策略
func (p *Pool) LoadBalancingStrategy() domain.LoadBalancingStrategy {
	return p.strategy
}

// 负载均衡策略
func (p *Pool) loadBalanced() (dataDir string, ip string, fingerPrint string) {
	switch p.strategy {
	case domain.Random:
		return utils.GetRandomly3(p.dataDirs, p.ips, p.fingerPrints)
	case domain.Round:
		return utils.GetByRound3(p.roundIdx.Add(1), p.dataDirs, p.ips, p.fingerPrints)
	default:
		return utils.GetRandomly3(p.dataDirs, p.ips, p.fingerPrints)
	}
}

// 获取该路径下的子文件夹路径
func (p *Pool) loadDataCatalog(basePath string) error {
	// 确保基础路径存在
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.MkdirAll(basePath, 0755)
		if err != nil {
			return fmt.Errorf("创建基础路径失败: %w", err)
		}
	}

	// 读取目录内容
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("读取数据目录失败: %w", err)
	}

	// 限制子目录扫描数量
	count := 0

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
	return nil
}
