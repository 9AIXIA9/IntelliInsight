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
	"sync/atomic"
	"time"
)

type Pool struct {
	units        chan domain.ResourceUnit //根据状态操作
	ips          []string                 //轮询
	dataPath     []string                 //轮询
	fingerPrints []string                 //轮询
	maxSize      int
	browserPath  string
	headless     bool
	counter      atomic.Uint32
}

func NewResourcePool(conf config.Resource) domain.ResourcePool {
	if len(conf.DataBasePath) == 0 {
		logx.Severef("资源池初始化失败：数据路径为空")
	}

	if len(conf.IPs) < 0 {
		logx.Severef("资源池初始化失败：IP数量过少")
	}

	if conf.MaxPoolSize < 0 {
		conf.MaxPoolSize = 5
	}

	if conf.InitialSize < 0 {
		conf.InitialSize = 2
	}

	pool := &Pool{
		units:        make(chan domain.ResourceUnit, conf.MaxPoolSize),
		ips:          conf.IPs,
		dataPath:     make([]string, conf.MaxPoolSize),
		fingerPrints: make([]string, conf.MaxPoolSize),
		maxSize:      conf.MaxPoolSize,
		browserPath:  conf.BrowserPath,
		headless:     conf.Headless,
	}

	if len(conf.FingerPrints) == 0 {
		pool.fingerPrints = domain.DefaultFingerPrints
	}

	//加载用户数据目录
	pool.loadDataCatalog(conf.DataBasePath)

	pool.initialize(conf.InitialSize)

	return pool
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
			p.dataPath[count] = dirPath
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
		p.dataPath[count] = dirPath
		count++
	}
}

// 提前加载资源
func (p *Pool) initialize(count int) {
	for i := 0; i < count; i++ {
		p.units <- p.createUnit()
	}
}

// 创建资源单元
func (p *Pool) createUnit() domain.ResourceUnit {
	id := p.counter.Add(1)

	return NewUnit(utils.GetByRound(id, p.dataPath), utils.GetByRound(id, p.ips), utils.GetByRound(id, p.fingerPrints), p.headless, p.browserPath)
}

func (p *Pool) GetResource(ctx context.Context) (domain.ResourceUnit, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case unit := <-p.units:
		return unit, nil
	}
}

func (p *Pool) ReleaseResource(unit domain.ResourceUnit) {
	threading.GoSafe(func() {
		p.units <- unit
	})
}

// RefreshResource 更新资源单元
func (p *Pool) RefreshResource(unit domain.ResourceUnit) domain.ResourceUnit {
	id := p.counter.Add(1)

	unit.Update(utils.GetByRound(id, p.dataPath), utils.GetByRound(id, p.ips), utils.GetByRound(id, p.fingerPrints))

	return unit
}

func (p *Pool) Close() {
	close(p.units)
}
