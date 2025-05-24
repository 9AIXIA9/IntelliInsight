package mongo

import (
	"context"
	"crawler/internal/domain"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository MongoDB仓库实现
type Repository struct {
	postCollection *mongo.Collection
	taskCollection *mongo.Collection
}

// NewRepository 创建MongoDB仓库实例
func NewRepository(client *mongo.Client, database string, postCollection, taskCollection string) domain.Repository {
	db := client.Database(database)
	return &Repository{
		postCollection: db.Collection(postCollection),
		taskCollection: db.Collection(taskCollection),
	}
}

// SavePosts 保存爬取到的帖子
func (r *Repository) SavePosts(ctx context.Context, taskID string, posts []*domain.Post) error {
	if len(posts) == 0 {
		return nil
	}

	var documents []interface{}
	for _, post := range posts {
		// 检查是否已存在(去重)
		filter := bson.M{"_id": post.ID}
		count, err := r.postCollection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}

		if count == 0 {
			// 确保每个post都有关联的taskID
			post.TaskID = taskID
			documents = append(documents, post)
		}
	}

	if len(documents) > 0 {
		_, err := r.postCollection.InsertMany(ctx, documents)
		return err
	}
	return nil
}

func (r *Repository) SaveTasks(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return errors.New("任务为空")
	}
	_, err := r.taskCollection.InsertOne(ctx, task)
	return err
}
