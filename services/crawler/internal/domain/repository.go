package domain

import (
	"context"
	"crawler/proto"
)

// Repository 定义仓库
type Repository interface {
	SaveTasks(ctx context.Context, task *Task) error
	SavePosts(ctx context.Context, taskID string, posts []*proto.PostItem) error
}
