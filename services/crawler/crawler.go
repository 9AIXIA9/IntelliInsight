package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"

	"crawler/internal/config"
	"crawler/internal/server"
	"crawler/internal/svc"
	"crawler/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/crawler.yaml", "the config file")

func main() {
	flag.Parse()

	//加载环境变量
	config.LoadEnv()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterCrawlerServiceServer(grpcServer, server.NewCrawlerServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 添加Consul服务注册
	if err := consul.RegisterService(c.ListenOn, c.Consul); err != nil {
		logx.Severe(fmt.Sprintf("注册服务到Consul失败: %v", err))
	}

	defer func() {
		s.Stop()
		ctx.CrawlerTaskQueue.Stop()
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
