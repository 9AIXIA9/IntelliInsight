package ctrl

import (
	"crawler/internal/domain"
	"crawler/internal/infra/utils/snowflake"
	"crawler/proto"
	"time"
)

// NewTask 创建新任务的工厂方法
func NewTask(request *proto.CrawlRequest) *domain.Task {
	return &domain.Task{
		ID:              snowflake.GenerateID(),
		ParentID:        "",
		Status:          domain.StatusPending,
		PostsCollected:  0,
		Err:             nil,
		StartTime:       time.Time{},
		EndTime:         time.Time{},
		Site:            request.Site,
		Keyword:         request.Keyword,
		PostCount:       request.PostCount,
		MinLikes:        request.MinLikes,
		CommentMinLikes: request.CommentMinLikes,
		CommentsPerPost: request.CommentsPerPost,
		IncludeComments: request.IncludeComments,
		IncludeImages:   request.IncludeImages,
	}
}

// 创建子任务
func createSubTask(parentTask *domain.Task, postCount uint64) *domain.Task {
	return &domain.Task{
		ID:              snowflake.GenerateID(),
		ParentID:        parentTask.ID,
		Status:          domain.StatusDivided,
		PostsCollected:  0,
		Err:             nil,
		StartTime:       parentTask.StartTime,
		EndTime:         parentTask.StartTime,
		Site:            parentTask.Site,
		Keyword:         parentTask.Keyword,
		PostCount:       postCount,
		MinLikes:        parentTask.MinLikes,
		CommentMinLikes: parentTask.CommentMinLikes,
		CommentsPerPost: parentTask.CommentsPerPost,
		IncludeComments: parentTask.IncludeComments,
		IncludeImages:   parentTask.IncludeImages,
	}
}
