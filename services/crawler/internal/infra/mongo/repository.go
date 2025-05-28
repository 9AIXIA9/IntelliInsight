package mongo

import (
	"context"
	"crawler/internal/domain"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

// SaveTask 保存任务，如果ID已存在则更新，不存在则插入
func (r *Repository) SaveTask(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return errors.New("任务为空")
	}

	// 处理错误字段
	var taskDoc interface{} = task
	if task.Err != nil {
		// 创建带错误字符串的文档副本
		doc := bson.D{}
		bytes, err := bson.Marshal(task)
		if err != nil {
			return err
		}

		if err := bson.Unmarshal(bytes, &doc); err != nil {
			return err
		}

		// 附加错误字符串
		doc = append(doc, bson.E{Key: "err", Value: task.Err.Error()})
		taskDoc = doc
	}

	// 使用upsert选项，存在则更新，不存在则插入
	filter := bson.M{"_id": task.ID}
	opts := options.Replace().SetUpsert(true)
	_, err := r.taskCollection.ReplaceOne(ctx, filter, taskDoc, opts)
	return err
}

func (r *Repository) UpdateParentTask(ctx context.Context, task *domain.Task) error {
	// 检查任务有效性
	if task == nil {
		return errors.New("任务为空")
	}

	// 检查任务是否有父任务
	if task.ParentID == "" {
		return errors.New("该任务没有父任务")
	}

	// 更新父任务：减少等待子任务数并增加帖子收集数
	filter := bson.M{"_id": task.ParentID}
	update := bson.M{
		"$inc": bson.M{
			"wait_sub_count":  -1,
			"posts_collected": task.PostsCollected,
		},
	}

	// 获取更新后的文档以检查wait_sub_count是否为0
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var result struct {
		WaitSubCount uint64 `bson:"wait_sub_count"`
	}

	err := r.taskCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("未找到父任务: %s", task.ParentID)
		}
		return fmt.Errorf("更新父任务进度失败: %w", err)
	}

	// 如果等待子任务数归零，则将任务标记为完成状态
	if result.WaitSubCount == 0 {
		completeUpdate := bson.M{
			"$set": bson.M{
				"status":   domain.StatusCompleted,
				"end_time": time.Now(),
			},
		}

		_, err = r.taskCollection.UpdateOne(ctx, filter, completeUpdate)
		if err != nil {
			return fmt.Errorf("标记父任务为完成状态失败: %w", err)
		}
	}

	return nil
}

// SavePosts 保存爬取到的帖子，返回新增帖子和评论
func (r *Repository) SavePosts(ctx context.Context, taskID domain.TaskID, posts []*domain.Post) error {
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
func (r *Repository) SavePostsAndComments(ctx context.Context, taskID domain.TaskID, posts []*domain.Post) error {
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
func (r *Repository) SaveComments(ctx context.Context, taskID domain.TaskID, postID string, comments []*domain.Comment) error {
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
