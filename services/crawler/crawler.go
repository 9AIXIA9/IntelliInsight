package main

import (
	"context"
	"crawler/internal/control"
	"crawler/internal/infra/utils/snowflake"
	"flag"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"time"

	"crawler/internal/config"
	"crawler/internal/server"
	"crawler/internal/svc"
	"crawler/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/crawler.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载环境变量
	config.LoadEnv()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.MustNewServiceContext(&c)

	snowflake.MustInit(c.SnowflakeID.StartTime, c.SnowflakeID.MachineNode)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterCrawlerServiceServer(grpcServer, server.NewCrawlerServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	control.Listen(&WrappedServer{
		conf:   &c,
		ctx:    ctx,
		server: s,
	})
}

type WrappedServer struct {
	conf   *config.Config
	ctx    *svc.ServiceContext
	server *zrpc.RpcServer
}

func (s *WrappedServer) Start() {
	// 添加Consul服务注册
	if err := consul.RegisterService(s.conf.ListenOn, s.conf.Consul); err != nil {
		control.LogSeveref("注册服务到Consul失败: %v", err)
		return
	}
	logx.Infof("成功注册服务到Consul")

	logx.Infof("启动RPC服务在：%s...\n", s.conf.ListenOn)

	s.server.Start()
}

func (s *WrappedServer) Stop() {
	logx.Info("收到退出信号，开始执行优雅关机...")

	// 创建一个带超时的上下文，确保关闭不会无限等待
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行优雅关机操作
	logx.Info("正在停止任务队列...")
	s.ctx.TaskQueue.Stop(stopCtx)
	logx.Info("任务队列已停止")

	//关闭服务器
	logx.Infof("关闭服务器...")
	s.server.Stop()
}
