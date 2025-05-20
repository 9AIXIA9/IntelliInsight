package domain

import (
	"crawler/proto"
)

const (
	defaultCommentsPerPost   = 10
	defaultRepliesPerComment = 5
)

// Crawler 执行爬虫
type Crawler interface {
	// CollectPostLinks 根据 minLikes 筛选网页中的postURL
	CollectPostLinks(resource ResourceUnit, site Site, keyword string, count int32, minLikes int32) (postLinks []string, err error)
	// CollectPostDetail 根据选项选择爬取帖子详情资源
	CollectPostDetail(resource ResourceUnit, site Site, postURL string, opts ...*CollectPostDetailOption) (*proto.PostItem, error)
}

type CollectPostDetailOption struct {
	IncludeComments   bool
	IncludeImages     bool
	CommentsPerPost   int32
	RepliesPerComment int32
}

func (opt *CollectPostDetailOption) WithDefault() *CollectPostDetailOption {
	if opt.CommentsPerPost < 0 {
		opt.CommentsPerPost = defaultCommentsPerPost
	}

	if opt.RepliesPerComment < 0 {
		opt.RepliesPerComment = defaultRepliesPerComment
	}

	return opt
}
