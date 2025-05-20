package mongo

import (
	"context"
	"crawler/internal/domain"
	"errors"
	"github.com/zeromicro/go-zero/core/stores/mon"

	"crawler/proto"
)

// Repository MongoDB仓库实现
type Repository struct {
	postModel *mon.Model
	taskModel *mon.Model
}

// NewRepository 创建MongoDB仓库实例
func NewRepository(postModel *mon.Model, taskModel *mon.Model) domain.Repository {
	return &Repository{postModel: postModel, taskModel: taskModel}
}

// SavePosts 保存爬取到的帖子
func (r *Repository) SavePosts(ctx context.Context, taskID string, posts []*proto.PostItem) error {
	if len(posts) == 0 {
		return nil
	}

	// 为每个帖子添加任务ID
	var documents []interface{}
	for _, post := range posts {
		doc := map[string]interface{}{
			"task_id": taskID,
			"post":    post,
		}
		documents = append(documents, doc)
	}

	_, err := r.postModel.InsertMany(ctx, documents)
	return err
}

func (r *Repository) SaveTasks(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return errors.New("任务为空")
	}

	_, err := r.taskModel.InsertOne(ctx, task)
	return err
}
