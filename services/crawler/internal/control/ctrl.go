package control

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/threading"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	quit = make(chan os.Signal) // 立即初始化通道
	wg   sync.WaitGroup
)

// LogSevere 严重错误处理
func LogSevere(format string, args ...interface{}) {
	logx.ErrorStackf(format, args...)

	// 触发优雅关机信号
	logx.Infof("触发关闭信号...")

	// 非阻塞地发送信号
	select {
	case quit <- syscall.SIGTERM:
		logx.Info("已发送关闭信号")
	default:
		// 如果没有监听者或通道已满，直接退出
		logx.Info("无法发送关闭信号，直接退出")
		os.Exit(1)
	}
}

func Listen(services ...service.Service) {

	// 确保通道已被初始化并监听信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动所有服务
	for _, s := range services {
		wg.Add(1)
		threading.GoSafe(func() {
			defer wg.Done()
			s.Start()
		})
	}

	// 等待退出信号
	<-quit

	// 停止所有服务
	for _, s := range services {
		s.Stop()
	}

	wg.Wait()
	logx.Info("所有服务已停止，程序即将退出")
	os.Exit(1)
}
