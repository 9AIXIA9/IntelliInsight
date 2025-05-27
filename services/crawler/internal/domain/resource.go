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
	Refresh(dataDir string, ip string, fingerPrint string) (ResourceUnit, error)
	CheckHealth() error
	Close()
}

// ResourcePool 资源池接口
type ResourcePool interface {
	Get(ctx context.Context) (ResourceUnit, error)
	Put(ResourceUnit)
	Refresh(ResourceUnit) (ResourceUnit, error)
	LoadBalancingStrategy() LoadBalancingStrategy
	AsyncHealthCheck(interval time.Duration)
	Close(ctx context.Context)
}

type LoadBalancingStrategy int

const (
	Random LoadBalancingStrategy = iota
	Round
)
