package domain

import (
	"time"
)

// Post 帖子模型
type Post struct {
	ID       string
	Title    string
	Poster   string
	Time     time.Time
	Location string
	Link     string
	Content  string
	Tags     []string

	LikeCount    uint64
	CommentCount uint64
	CollectCount uint64

	ImageURLs []string
	Comments  []*Comment
}

// Comment 评论模型
type Comment struct {
	ID        string
	Commenter string
	Time      time.Time
	Location  string
	Content   string

	LikeCount  uint64
	ReplyCount uint64

	//Replies []*Comment
}

// 暂时降低难度 暂不实现回复
//因为回复大部分都是在闲聊 性价比太低了
//// Reply 回复模型
//type Reply struct {
//	ID       string
//	Replier  string
//	Time     time.Time
//	Location string
//	Content  string
//
//	LikeCount uint64
//}
