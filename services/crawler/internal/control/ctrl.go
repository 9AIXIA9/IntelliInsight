package control

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"os"
	"os/signal"
	"syscall"
)

var (
	quit = make(chan os.Signal) // 立即初始化通道
)

type Unit interface {
	Start()
	Stop()
}

// LogSeveref 严重错误处理
func LogSeveref(format string, args ...interface{}) {
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

func Listen(units ...Unit) {

	// 确保通道已被初始化并监听信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动所有服务
	for _, s := range units {
		threading.GoSafe(func() {
			s.Start()
		})
	}

	// 等待退出信号
	<-quit

	// 停止所有服务
	for _, u := range units {
		u.Stop()
	}

	logx.Info("所有服务已停止，程序即将退出")
}
