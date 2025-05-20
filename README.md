# 项目概述 - IntelliInsight 智析

## 项目架构

本项目是一个分布式爬虫与数据分析系统，采用微服务架构：

- **前端/分析平台**：Python FastAPI 服务
- **爬虫核心**：Go 服务
- **通信方式**：gRPC
- **数据存储**：MongoDB
- **服务发现与注册**：Consul

## 技术栈

Go服务专注于高效爬取
Python服务专注于数据展示和分析
MongoDB作为中间层实现数据交换

- **Python 端**：
- FastAPI 框架
- Motor (异步 MongoDB 驱动)
- gRPC 客户端
- Pydantic 数据验证
- Uvicorn 服务器
- 基础爬虫（降级方案）

- **Go 端**：
- 高性能爬虫核心
- gRPC 服务端

## 系统特点

- 高性能：Go 实现的爬虫核心
- 易用性：Python FastAPI 提供友好的 REST API 和 Swagger 文档
- 可扩展：微服务架构便于横向扩展
- 异步操作：全链路异步支持高并发请求

## 主要功能

1. 创建爬虫任务：指定关键词、页数等参数
2. 查询爬虫状态：实时监控爬虫进度
3. 获取爬取数据：分页查询
4. 智能数据分析
5. 降级方案（当gRPC服务不可用时，使用降级方案保证系统的可用性）
6. 去重防止爬取相同内容（布隆过滤器）

## 爬虫设计

1. 分析解构小红书网页组成，抽象成数据结构（protobuf文件）

2. 并发爬虫设计：
	- 优势：当sleep或者阻塞时（尤其在反爬机制实现时，go能够高效并发，所以适合做并发爬虫），及时让出CPU以及其余资源的占有，提高资源利用率
	- 生产者 收到任务后模拟搜索，读取所需信息后，将解析到的URL加入全局任务队列
	- 消费者 监听全局任务队列，收到任务时被唤醒，访问IP池获取IP代理，访问URL爬取数据 保证数量不要超过限制
	- 动态更换用户代理和IP代理
3. 常规爬虫设计：单线程爬取网页，任务一个一个解决
4. 高安全模式：
   浏览器实例:IP = 1:1 固定绑定
   资源更换条件: 仅当检测到封禁

## 存储设计

```api
type Task struct {
    ID string `bson:"_id"`
    Keyword string `bson:"keyword"`
    PostCount int32 `bson:"post_count"`
    IncludeComments bool `bson:"include_comments"`
    MinLikes int32 `bson:"min_likes"`
    CommentsPerPost int32 `bson:"comments_per_post"`
    RepliesPerComment int32 `bson:"replies_per_comment"`
    IncludeImages bool `bson:"include_images"`
    Status proto.TaskStatus `bson:"status"`
    Progress float32 `bson:"progress"`
    ItemsCollected int32 `bson:"items_collected"`
    CommentsCollected int32 `bson:"comments_collected"`
    RepliesCollected int32 `bson:"replies_collected"`
    Message string `bson:"message"`
    StartTime time.Time `bson:"start_time"`
    EndTime time.Time `bson:"end_time,omitempty"`
    ErrorMessage string `bson:"error_message,omitempty"`
    ElapsedTime int64 `bson:"elapsed_time"` // 单位：秒
}

type PostItem struct {
    PostId string                       // 帖子ID
    Title string                        // 标题 --> div.note-content 下的 title
    Content string                      // 内容 --> note-content 下的 note-text
    Author string                       // 作者 -->  span.username 选择第一个
    Likes string                        // 点赞数  span.like-wrapper 下的 count
    PublishTime int64                   // 发布时间戳 class date 今天 07:00 湖南 前半段
    Images []string                     // 图片URL列表  .div.img-container 下的img
    Comments []*Comment                 // 评论列表 .div.list-container 下
    Tags []string                       // 标签列表 --> div.note-content 下的 desc
    Location string                     // 位置信息 class date 后半段
    Collects string                     // 收藏数 span.collect-wrapper 下的 count
    Chats string                        // 评论数 span.chat-wrapper 下的 count
}

type Comment struct {
    CommentId string                        // 评论ID class comment-item 的id就是评论id
    Content string                          // 评论内容 content 下的 note-text
    Author string                           // 评论作者 author 但是是一个链接 a 可从中获取名字
    Likes string                            // 评论点赞数 like 下的 count
    CommentTime int64                       // 评论时间戳 date （格式为04-16）
    CommentLocation string                  // 评论位置 location （格式为湖南） 可能不存在
    Replies []*Comment                      // 回复列表 reply-container
    RepliesCount string                     // 回复数 reply 下的 count
}

```