package domain

import (
	"context"
	"crawler/proto"
)

// TaskQueue 管理生产者消费者 充当协调器
type TaskQueue interface {
	AddTask(request *proto.CrawlRequest) (ID string, err error) // 添加新任务并返回任务ID
	Dispatcher()                                                // 负责分配任务给工作协程
	ProcessTask(task *Task)                                     // 处理单个任务的逻辑
	Stop(ctx context.Context)                                   // 停止
}
