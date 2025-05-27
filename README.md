# 项目概述 - IntelliInsight 智析

## 项目架构

本项目是一个分布式爬虫与数据分析系统，采用微服务架构：

- **前端/分析平台**：Python FastAPI 服务
- **爬虫核心**：Go 服务
- **通信方式**：gRPC
- **数据存储**：MongoDB
- **数据过滤**：Redis
- **服务发现与注册**：Consul

## 技术栈

Go服务专注于高效爬取
Python服务专注于数据展示
MongoDB作为中间层实现数据交换

- **Python 端**：
- FastAPI 框架
- gRPC 客户端
- Uvicorn 服务器

- **Go 端**：
- 高性能并发爬虫核心
- gRPC 服务端
- MongoDB数据存储
- Redis布隆过滤器过滤数据

## 系统特点

- 高性能：Go 实现的爬虫核心
- 易用性：Python FastAPI 提供友好的 REST API 和 Swagger 文档
- 可扩展：微服务架构便于横向扩展
- 异步操作：全链路异步支持高并发请求

## 主要功能

1. 创建爬虫任务：指定关键词、页数等参数
2. 降级方案（当gRPC服务不可用时，使用降级方案保证系统的可用性）
3. 去重防止爬取相同内容（布隆过滤器）

## 爬虫设计

1. 分析解构小红书网页组成，整理成数据结构（domain文件）

   ```go
	  package domain
   
   import (
	   "time"
   )
   
   // Task 爬虫任务模型
   type Task struct {
	   ID             string    `bson:"_id"`
	   StartTime      time.Time `bson:"start_time"`
	   EndTime        time.Time `bson:"end_time"`
	   Status         string    `bson:"status"`
	   PostsCollected uint32    `bson:"posts_collected"`
	   Err            error     `bson:"err,omitempty"`
	   
	   // CrawlRequest字段展平
	   Site            proto.Site `bson:"site"`
	   Keyword         string     `bson:"keyword"`
	   PostCount       uint64     `bson:"post_count"`
	   MinLikes        uint64     `bson:"min_likes"`
	   CommentMinLikes uint64     `bson:"comment_min_likes"`
	   CommentsPerPost uint64     `bson:"comments_per_post"`
	   IncludeComments bool       `bson:"include_comments"`
	   IncludeImages   bool       `bson:"include_images"`
   }
   
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
	   Comments  []*Comment `bson:"-"`
   }
   
   // Comment 评论模型
   type Comment struct {
	   ID        string    `bson:"_id"`
	   TaskID    string    `bson:"task_id"`
	   PostID    string    `bson:"post_id"`
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
   ```

2. 并发爬虫设计：
	- 优势：当sleep或者阻塞时（尤其在反爬机制实现时，go能够高效并发，所以适合做并发爬虫），及时让出CPU以及其余资源的占有，提高资源利用率
	- 生产者 收到任务后模拟搜索，读取所需信息后，将解析到的URL加入全局任务队列
	- 消费者 监听全局任务队列，收到任务时被唤醒，访问IP池获取IP代理，访问URL爬取数据 保证数量不要超过限制
	- 动态更换用户代理和IP代理
3. 常规爬虫设计：单线程爬取网页，任务一个一个解决
4. 高安全模式：
   浏览器实例:IP = 1:1 固定绑定
   资源更换条件: 仅当检测到封禁

## todo

1. 降级方案
2. 扩展方向

## 扩展方向

- ~~引入Consul进行服务注册和发现，并实时进行健康检查~~
- ~~引入Redis布隆过滤器，对已爬取的页面不再重复爬取~~
- ~~重新设计存储模式，发挥MongoDB的特性（横向扩展）~~
- 扩展Python-APP的交互功能
- ~~完成手机验证码自动登录 -> 无法实现~~
  实现此功能需要 虚拟手机号API+自动化脚本 或 使用本机号码 + 转发短信服务 风险过高
  使用手动扫码替代此功能
- 分治处理多站点及大任务
- 架构宏观扩展
- 接入metrics端点，暴露任务数、爬取速度等指标

### 分治方案

- 策略
	1. 规模 · 根据任务大小进行分治 如数据量为1000的任务分为5个200的小任务，分配到协程中运行
	2. 来源 · 根据任务需求爬取不同站点 如一个协程爬取小红书一个爬取知乎
- 设计
	1. TaskQueue中的更改
		- 添加AddSubTask()方法 - 作为向内提供的分治兼容函数
		- Dispatcher()任务调度器 - 添加一个分支用于处理子任务
		- 重构ProcessTask - 抽离公用逻辑作为函数而不是方法 能够完成SubTask和Task子任务和完整的任务
	2.

### 架构宏观扩展方向

1. 接入metrics端点，暴露任务数、爬取速度等指标
2. 协程 -> 爬虫微服务
3. chan通道 -> 任务队列
4. Redis和MongoDB单点 -> Redis和MongoDB集群