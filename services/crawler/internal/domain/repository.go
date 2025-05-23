package domain

import (
	"context"
)

// Repository 定义仓库
type Repository interface {
	SaveTasks(ctx context.Context, task *Task) error
	SavePosts(ctx context.Context, taskID string, posts []*Post) error
}
