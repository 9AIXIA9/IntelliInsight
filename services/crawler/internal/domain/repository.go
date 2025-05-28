package domain

import (
	"context"
)

// Repository 定义仓库
type Repository interface {
	SaveTask(ctx context.Context, task *Task) error
	UpdateParentTask(ctx context.Context, task *Task) error
	SavePostsAndComments(ctx context.Context, taskID TaskID, posts []*Post) error
	SavePosts(ctx context.Context, taskID TaskID, posts []*Post) error
	SaveComments(ctx context.Context, taskID TaskID, postID string, comments []*Comment) error
}
