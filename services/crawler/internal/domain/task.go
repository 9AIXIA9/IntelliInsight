package domain

import (
	"crawler/proto"
	"time"
)

// Task 爬虫任务模型
type Task struct {
	ID             string
	Request        *proto.CrawlRequest
	PostsCollected uint32
	StartTime      time.Time
	EndTime        time.Time
	Status         string
	Err            error
}

// TaskStatus 任务状态
type TaskStatus int

const (
	StatusCompleted TaskStatus = iota
	StatusFailed
	StatusRunning
	StatusPending
)

func (s TaskStatus) String() string {
	switch s {
	case StatusCompleted:
		return "已完成"
	case StatusFailed:
		return "失败"
	case StatusRunning:
		return "运行中"
	case StatusPending:
		return "等待中"
	default:
		return "未知状态"
	}
}
