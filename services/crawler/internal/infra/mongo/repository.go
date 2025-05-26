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
	postCollection    *mongo.Collection
	commentCollection *mongo.Collection
	taskCollection    *mongo.Collection
}

// NewRepository 创建MongoDB仓库实例
func NewRepository(client *mongo.Client, database string, postCollection, commentCollection, taskCollection string) domain.Repository {
	db := client.Database(database)
	return &Repository{
		postCollection:    db.Collection(postCollection),
		commentCollection: db.Collection(commentCollection),
		taskCollection:    db.Collection(taskCollection),
	}
}

func (r *Repository) SaveTask(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return errors.New("任务为空")
	}

	// 检查任务是否已存在
	filter := bson.M{"_id": task.ID}
	count, err := r.taskCollection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("任务ID已存在")
	}

	_, err = r.taskCollection.InsertOne(ctx, task)
	return err
}

// SavePosts 保存爬取到的帖子，返回新增帖子和评论
func (r *Repository) SavePosts(ctx context.Context, taskID string, posts []*domain.Post) error {
	if len(posts) == 0 {
		return nil
	}

	var documents []interface{}
	var postIDs []string

	// 收集所有帖子ID用于批量查询
	for _, post := range posts {
		post.TaskID = taskID
		postIDs = append(postIDs, post.ID)
	}

	// 批量查询已存在的帖子
	filter := bson.M{"_id": bson.M{"$in": postIDs}}
	cursor, err := r.postCollection.Find(ctx, filter)
	if err != nil {
		return err
	}

	existingPosts := make(map[string]bool)
	var existingPostsList []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &existingPostsList); err != nil {
		return err
	}

	for _, p := range existingPostsList {
		existingPosts[p.ID] = true
	}

	// 处理不存在的帖子
	for _, post := range posts {
		if !existingPosts[post.ID] {
			// 清空帖子中的评论字段前创建副本
			postCopy := *post
			postCopy.Comments = nil
			documents = append(documents, &postCopy)
		}
	}

	// 插入新帖子
	if len(documents) > 0 {
		_, err := r.postCollection.InsertMany(ctx, documents)
		if err != nil {
			return err
		}
	}

	return nil
}

// SavePostsAndComments 保存爬取到的帖子和评论
func (r *Repository) SavePostsAndComments(ctx context.Context, taskID string, posts []*domain.Post) error {
	if len(posts) == 0 {
		return nil
	}

	// 先保存所有帖子
	if err := r.SavePosts(ctx, taskID, posts); err != nil {
		return err
	}

	// 批量查询已存在的帖子，用于确定哪些帖子的评论需要保存
	var postIDs []string
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	filter := bson.M{"_id": bson.M{"$in": postIDs}}
	cursor, err := r.postCollection.Find(ctx, filter)
	if err != nil {
		return err
	}

	existingPosts := make(map[string]bool)
	var existingPostsList []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &existingPostsList); err != nil {
		return err
	}

	for _, p := range existingPostsList {
		existingPosts[p.ID] = true
	}

	// 保存所有评论
	for _, post := range posts {
		if len(post.Comments) > 0 {
			if err := r.SaveComments(ctx, taskID, post.ID, post.Comments); err != nil {
				return err
			}
		}
	}

	return nil
}

// SaveComments 保存爬取到的评论到独立集合
func (r *Repository) SaveComments(ctx context.Context, taskID, postID string, comments []*domain.Comment) error {
	if len(comments) == 0 {
		return nil
	}

	var documents []interface{}
	for _, comment := range comments {
		comment.PostID = postID
		comment.TaskID = taskID

		// 检查评论是否已存在(去重)
		filter := bson.M{"_id": comment.ID}
		count, err := r.commentCollection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}

		if count == 0 {
			documents = append(documents, comment)
		}
	}

	if len(documents) > 0 {
		_, err := r.commentCollection.InsertMany(ctx, documents)
		return err
	}
	return nil
}
