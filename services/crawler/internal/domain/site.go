package domain

import "time"

// Site 接口定义站点特定的行为和知识
type Site interface {
	GetName() string
	GetSearchURL(keyword string) string
	GetBaseURL() string

	GetLoginCardSelector() string

	GetPostCardSelector() string
	GetPostLinkSelector() string
	GetPostPageLikeCountSelector() string

	GetPostTitleSelector() string
	GetPostContentSelector() string
	GetPosterSelector() string
	GetPostTagsSelector() string
	GetPostDetailLikeCountSelector() string
	GetPostDetailCommentCountSelector() string
	GetPostDetailCollectCountSelector() string
	GetImageURLSelector() string
	GetTimeSelector() string
	GetLocationSelector() string

	GetCommentItemSelector() string
	GetCommenterSelector() string
	GetCommentTimeSelector() string
	GetCommentLocationSelector() string
	GetCommentContentSelector() string
	GetCommentLikeCountSelector() string
	GetCommentReplyCountSelector() string

	ParsePostLinks(html string, filter Filter, minLikes uint64) ([]string, error)
	ParsePostPage(html string, includeImages bool) (*Post, error)
	ParseComments(html string, count uint64, minLikes uint64) ([]*Comment, error)

	ParseNumber(numStr string) uint64
	ParseLocation(locationStr string) string
	ParseTime(timeStr string) (time.Time, error)
	ParsePostIDFromURL(url string) string
	ParseCommentID(idStr string) string

	RequireLogin(html string) (bool, error)
	Login(browser Browser) error
}
