package domain

import (
	"time"
)

// Post 帖子模型
type Post struct {
	ID       string    `bson:"_id"`
	TaskID   string    `bson:"task_id"`
	Title    string    `bson:"title"`
	Poster   string    `bson:"poster"`
	Time     time.Time `bson:"time"`
	Location string    `bson:"location"`
	Link     string    `bson:"link"`
	Content  string    `bson:"content"`
	Tags     []string  `bson:"tags"`

	LikeCount    uint64 `bson:"like_count"`
	CommentCount uint64 `bson:"comment_count"`
	CollectCount uint64 `bson:"collect_count"`

	ImageURLs []string   `bson:"image_urls"`
	Comments  []*Comment `bson:"comments"`
}

// Comment 评论模型
type Comment struct {
	ID        string    `bson:"_id"`
	Commenter string    `bson:"commenter"`
	Time      time.Time `bson:"time"`
	Location  string    `bson:"location"`
	Content   string    `bson:"content"`

	LikeCount  uint64 `bson:"like_count"`
	ReplyCount uint64 `bson:"reply_count"`

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
