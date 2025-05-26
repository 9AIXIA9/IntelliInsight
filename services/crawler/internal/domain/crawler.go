package domain

const (
	defaultCommentsPerPost = 10
)

// Crawler 执行爬虫
type Crawler interface {
	// CollectPostLinks 根据 minLikes 筛选网页中的帖子链接
	CollectPostLinks(resource Browser, filter Filter, site Site, keyword string, count uint64, minLikes uint64) (links []string, err error)
	// CollectPostDetail 根据选项选择爬取帖子详情资源
	CollectPostDetail(resource Browser, site Site, postURL string, opts ...*CollectPostDetailOption) (*Post, error)
}

type CollectPostDetailOption struct {
	IncludeComments  bool
	IncludeImages    bool
	CommentsPerPost  uint64
	MinLikes         uint64
	CommentsMinLikes uint64
}

func (opt *CollectPostDetailOption) WithDefault() *CollectPostDetailOption {
	if opt.CommentsPerPost < 0 {
		opt.CommentsPerPost = defaultCommentsPerPost
	}

	if opt.MinLikes < 0 {
		opt.MinLikes = 0
	}

	return opt
}
