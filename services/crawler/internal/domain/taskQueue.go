package domain

import (
	"crawler/proto"
	"time"
)

// TaskQueue 管理生产者消费者 充当协调器
type TaskQueue interface {
	AddTask(request *proto.CrawlRequest) (ID string, err error) // 添加新任务并返回任务ID
	Dispatcher()                                                // 负责分配任务给工作协程
	ProcessTask(task *Task)                                     // 处理单个任务的逻辑
	Stop()                                                      // 停止
}

// Task 表示一个爬虫任务
type Task struct {
	ID             string
	Request        *proto.CrawlRequest
	PostsCollected int32
	StartTime      time.Time
	EndTime        time.Time
	Status         TaskStatus
	Err            error
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusCompleted TaskStatus = iota
	TaskStatusFailed
	TaskStatusRunning
	TaskStatusPending
)
