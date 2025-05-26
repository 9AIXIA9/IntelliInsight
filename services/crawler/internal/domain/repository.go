package domain

import (
	"context"
)

// Repository 定义仓库
type Repository interface {
	SaveTask(ctx context.Context, task *Task) error
	SavePostsAndComments(ctx context.Context, taskID string, posts []*Post) error
	SavePosts(ctx context.Context, taskID string, posts []*Post) error
	SaveComments(ctx context.Context, taskID, postID string, comments []*Comment) error
}
