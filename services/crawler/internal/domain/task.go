package domain

import (
	"crawler/proto"
	"time"
)

// Task 爬虫任务模型
type Task struct {
	ID             string              `bson:"_id"`
	Request        *proto.CrawlRequest `bson:"request"`
	PostsCollected uint32              `bson:"posts_collected"`
	StartTime      time.Time           `bson:"start_time"`
	EndTime        time.Time           `bson:"end_time"`
	Status         string              `bson:"status"`
	Err            error               `bson:"err,omitempty"`
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
