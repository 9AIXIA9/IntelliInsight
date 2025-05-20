package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf

	MongoDB MongoDB

	Consul consul.Conf

	TaskQueue TaskQueue

	Resource Resource
}

type MongoDB struct {
	URI            string
	Database       string
	PostCollection string
	TaskCollection string
}
type TaskQueue struct {
	WorkQueueSize int
	MaxWorkers    int
}

type Resource struct {
	DataBasePath string   //存储数据的目录
	IPs          []string //IPs
	FingerPrints []string //指纹

	BrowserPath string //浏览器路径
	Headless    bool   //是否开启可视化 默认关闭
	MaxPoolSize int    // 最大资源实例数 默认 5
	InitialSize int    // 初始资源实例数 默认 2
}
