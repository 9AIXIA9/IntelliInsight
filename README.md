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

``` prototext
// 帖子模型
message PostItem {
  string post_id = 1;                 // 帖子ID https://www.xiaohongshu.com/explore/676425e6000000000b0164ab?xsec_token=
  string title = 2;                   // 标题 --> div.note-content 下的 title
  string content = 3;                 // 内容 --> note-content 下的 note-text
  string author = 4;                  // 作者 -->  span.username 选择第一个
  int32 likes = 5;                    // 点赞数  span.like-wrapper 下的 count
  int64 post_time = 6;                // 帖子时间戳 class date 今天 07:00 湖南 前半段  处理"今天"、"昨天"等相对时间 1天前 2天前 昨天 17:17 13小时前 03-09 -> 代表今年三月9号 2024-10-25
  repeated string images = 7;         // 图片URL列表  .div.img-container 下的img
  repeated Comment comments = 8;      // 评论列表 .div.list-container 下
  repeated string tags = 9;           // 标签列表 --> div.note-content 下的 desc
  string location = 10;               // 位置信息 class date 后半段
  int32 collects = 11;                // 收藏数 span.collect-wrapper 下的 count
  int32 chats = 12;                   // 评论数 span.chat-wrapper 下的 count
  string url = 13;                    // 帖子链接 note-item 下的 a href 是绝对路径
}

// 评论模型
// 不用点击评论 自动会出来评论 同样也是无限下滑流的设计
// replies 自动会浮现第一条回复 点击.div.show-more后就会加载更多回复
message Comment {
  string comment_id = 1;            // 评论ID  div.comment-item 的id就是评论id
  string content = 2;               // 评论内容 content 下的 note-text
  string author = 3;                // 评论作者 author 但是是一个链接 a 可从中获取名字
  int32 likes = 4;                  // 评论点赞数 div.like 下的 count
  int64 comment_time = 5;           // 评论时间戳 date （格式为04-16）
  string comment_location = 6;      // 评论位置 location （格式为湖南） 可能不存在
  repeated Comment replies = 7;     // 回复列表 reply-container
  int32 replies_count = 8;          // 回复数 div.reply 下的 count
}
```