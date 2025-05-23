package domain

const (
	defaultCommentsPerPost   = 10
	defaultRepliesPerComment = 5
)

// Crawler 执行爬虫
type Crawler interface {
	// CollectPostLinks 根据 minLikes 筛选网页中的帖子链接
	CollectPostLinks(resource Browser, site Site, keyword string, count uint64, minLikes uint64) (links []string, err error)
	// CollectPostDetail 根据选项选择爬取帖子详情资源
	CollectPostDetail(resource Browser, site Site, postURL string, opts ...*CollectPostDetailOption) (*Post, error)
}

type CollectPostDetailOption struct {
	IncludeComments   bool
	IncludeImages     bool
	CommentsPerPost   uint64
	RepliesPerComment uint64
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
