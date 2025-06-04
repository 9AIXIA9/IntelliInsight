package domain

import (
	"context"
)

// TaskQueue 管理生产者消费者 充当协调器
type TaskQueue interface {
	Dispatcher()              // 任务调度器 用于分配任务给协程
	AddTask(task *Task) error // 添加新任务
	ProcessTask(task *Task)   // 处理单个任务的逻辑
	Stop(ctx context.Context) // 停止
	Divider                   //分治处理器
	Merger                    //合并处理器
}

type Divider interface {
	NeedToDivide(task *Task) bool  //判断是否需要分治
	DivideTask(task *Task) []*Task //返回小任务
}

type Merger interface {
	UpdateParentTaskProgress(task *Task) error //更新父任务的进度
}
