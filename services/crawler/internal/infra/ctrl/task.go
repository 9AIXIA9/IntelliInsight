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
		Info: &domain.TaskInfo{
			ID:             snowflake.GenerateID(),
			ParentID:       "",
			Status:         domain.StatusPending.String(),
			PostsCollected: 0,
			Err:            nil,
			StartTime:      time.Time{},
			EndTime:        time.Time{},
		},
		Request: &domain.TaskRequest{
			Site:            request.Site,
			Keyword:         request.Keyword,
			PostCount:       request.PostCount,
			MinLikes:        request.MinLikes,
			CommentMinLikes: request.CommentMinLikes,
			CommentsPerPost: request.CommentsPerPost,
			IncludeComments: request.IncludeComments,
			IncludeImages:   request.IncludeImages,
		},
	}
}
