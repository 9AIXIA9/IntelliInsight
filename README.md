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
Python服务专注于数据展示
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

## 数据模型设计

- 存储模型

```go
package domain

import (
	"crawler/proto"
	"time"
)

//Post 帖子模型
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

//Task 爬虫任务模型
type Task struct {
	ID             string
	Request        *proto.CrawlRequest
	PostsCollected uint32
	StartTime      time.Time
	EndTime        time.Time
	Status         string
	Err            error
}

 ```

## 扩展方向

- ~~引入Consul进行服务注册和发现，并实时进行健康检查~~
- 引入Redis布隆过滤器，对已爬取的页面不再重复爬取
- 重新设计存储模式，使数据更易懂
- 扩展Python-APP的交互功能
- 完成手机验证码自动登录

