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
		ID:             snowflake.GenerateID(), // 需要实现生成ID的函数
		StartTime:      time.Now(),
		Status:         domain.StatusPending.String(),
		PostsCollected: 0,

		// 将request字段展平
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
