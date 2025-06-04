# 系统启动指南

## 系统架构概述

该系统是一个分布式爬虫与数据分析平台，主要包括：

1. **Golang爬虫服务**：提供核心爬虫功能，使用资源池管理浏览器实例
2. **Python API服务**：提供RESTful API接口，供前端调用

## 前置要求

- MongoDB (v4.4+)
- Consul (服务发现，可选)
- Google Chrome (爬虫使用)
- Go 1.19+
- Python 3.9+

## 启动步骤

### 1. 配置环境

1. 复制环境示例文件并根据实际情况修改：

   ```bash
   # Golang服务
   cp .env.example
   
   # Python服务
   cp app/.env.example app/.env
   ```

2. 确保MongoDB已启动并可访问。

3. 添加一个.env文件到golang中，内容为

```
APP_ENV=example
```

### 2. 启动Golang爬虫服务

```bash
# 进入爬虫服务目录
cd services/crawler

# 构建并运行
go build -o crawler
./crawler
```

### 3. 启动Python API服务

```bash
# 进入API服务目录
cd app

# 安装依赖
pip install -r requirements.txt

# 启动服务
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## 验证服务

1. API服务文档访问：http://localhost:8000/docs
2. 可以通过API创建爬虫任务，监控任务状态，查询爬取结果

## 常见问题

1. **爬虫服务资源池错误**：
	- 检查Chrome浏览器路径是否正确
	- 确认IP代理配置是否有效

2. **服务发现失败**：
	- 检查Consul服务是否正常运行
	- 验证服务注册键值是否正确

3. **数据库连接问题**：
	- 确认MongoDB连接字符串正确
	- 检查数据库用户权限

4. **服务间通信失败**：
	- 确保gRPC服务端口已开放
	- 检查网络防火墙设置

## 系统功能说明

- 支持小红书平台爬虫
- 可按关键词进行爬取
- 支持帖子、评论数据爬取与持久化
- 提供任务分治机制处理大型爬虫任务
- 支持任务状态跟踪与管理