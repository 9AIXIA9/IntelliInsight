package config

import (
	"crawler/internal/domain"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"time"
)

type Config struct {
	zrpc.RpcServerConf

	Consul consul.Conf

	Redisx redis.RedisConf

	MongoDB MongoDB

	TaskQueue TaskQueue

	Resource Resource

	BloomFilter BloomFilter

	SnowflakeID SnowflakeID
}

type MongoDB struct {
	URI               string
	Database          string
	PostCollection    string
	CommentCollection string
	TaskCollection    string
}

type TaskQueue struct {
	WorkQueueSize    int
	MaxWorkers       int
	MaxTaskCacheSize int
	DivideThreshold  uint64
}

type Resource struct {
	DataBasePath string   //存储数据的目录
	IPs          []string //IPs
	FingerPrints []string //指纹

	BrowserPath           string                       //浏览器路径
	MaxPoolSize           int                          // 最大资源实例数
	InitialSize           int                          // 初始资源实例数
	Headless              bool                         //是否开启可视化
	HealthCheckInterval   time.Duration                //健康检查间隔
	LoadBalancingStrategy domain.LoadBalancingStrategy //负载均衡策略
}

type BloomFilter struct {
	Bits uint
	Key  string
}

type SnowflakeID struct {
	StartTime   time.Time
	MachineNode int64
}
