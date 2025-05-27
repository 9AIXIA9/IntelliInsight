package domain

import (
	"crawler/proto"
	"time"
)

type TaskID string

// Task 爬虫任务模型
type Task struct {
	// 任务基本信息
	ID             TaskID     `bson:"_id"`
	ParentID       TaskID     `bson:"parent_id,omitempty"` // 未分治任务ParentID为空
	Status         TaskStatus `bson:"status"`
	PostsCollected uint32     `bson:"posts_collected"`
	StartTime      time.Time  `bson:"start_time"`
	EndTime        time.Time  `bson:"end_time"`
	Err            error      `bson:"err,omitempty"`

	// 任务请求参数
	Site            proto.Site `bson:"site"`
	Keyword         string     `bson:"keyword"`
	PostCount       uint64     `bson:"post_count"`
	MinLikes        uint64     `bson:"min_likes"`
	CommentMinLikes uint64     `bson:"comment_min_likes"`
	CommentsPerPost uint64     `bson:"comments_per_post"`
	IncludeComments bool       `bson:"include_comments"`
	IncludeImages   bool       `bson:"include_images"`
}

// TaskStatus 任务状态
type TaskStatus int

const (
	StatusCompleted TaskStatus = iota
	StatusFailed
	StatusRunning
	StatusPending
	StatusDivided //分治
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
	case StatusDivided:
		return "分治处理中"
	default:
		return "未知状态"
	}
}
