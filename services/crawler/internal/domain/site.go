package domain

import "crawler/proto"

// Site 接口定义站点特定的行为和知识
type Site interface {
	GetName() string
	GetSearchURL(keyword string) string
	GetBaseURL() string

	GetPostCardSelector() string
	GetPostLinkSelector() string
	GetPostTitleSelector() string
	GetPostContentSelector() string
	GetAuthorSelector() string
	GetLikesSelector() string
	GetChatsSelector() string
	GetCollectsSelector() string
	GetImagesSelector() string
	GetTagsSelector() string
	GetTimeSelector() string
	GetLocationSelector() string

	ParsePostLinks(html string, minLikes int32) ([]string, error)
	ParsePostDetail(html string, includeImages bool) (*proto.PostItem, error)
	ParseComments(html string, count int32, repliesCount int32) ([]*proto.Comment, error)

	ParseLocation(locationStr string) string
	ParseTime(timeStr string) (int64, error)
	ParseNumber(numStr string) int32
	ParsePostIDFromURL(url string) string

	NeedsLogin(html string) bool
	Login() error
}
