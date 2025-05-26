package domain

import (
	"crawler/proto"
	"time"
)

// Task 爬虫任务模型
type Task struct {
	ID             string    `bson:"_id"`
	StartTime      time.Time `bson:"start_time"`
	EndTime        time.Time `bson:"end_time"`
	Status         string    `bson:"status"`
	PostsCollected uint32    `bson:"posts_collected"`
	Err            error     `bson:"err,omitempty"`

	// CrawlRequest字段展平
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
