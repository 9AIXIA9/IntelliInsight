package domain

import (
	"context"
	"time"
)

// ResourceUnit 表示一个完整的资源单元
type ResourceUnit interface {
	Browser() Browser
	IP() string
	Fingerprint() string
	DataDir() string
	Refresh(dataDir string, ip string, fingerPrint string)
	CheckHealth() bool
	Close()
}

// ResourcePool 资源池接口
type ResourcePool interface {
	GetResource(ctx context.Context) (ResourceUnit, error)
	ReleaseResource(ResourceUnit)
	RefreshResource(ResourceUnit) ResourceUnit
	AsyncHealthCheck(interval time.Duration)
	CheckUnitsHealth()
	LoadBalancingStrategy() LoadBalancingStrategy
	LoadBalancing() (dataDir string, ip string, fingerPrint string)
	Close()
}

type LoadBalancingStrategy int

const (
	Random LoadBalancingStrategy = iota
	Round
)
