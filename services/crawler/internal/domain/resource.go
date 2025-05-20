package domain

import "context"

// ResourceUnit 表示一个完整的资源单元
type ResourceUnit interface {
	Browser() Browser
	IP() string
	Fingerprint() string
	DataDir() string
	Update(dataDir string, ip string, fingerPrint string)
}

// ResourcePool 资源池接口
type ResourcePool interface {
	GetResource(ctx context.Context) (ResourceUnit, error)
	ReleaseResource(ResourceUnit)
	RefreshResource(ResourceUnit) ResourceUnit
	Close()
}

// DefaultFingerPrints todo 填补
var DefaultFingerPrints = []string{"demo"}
