import os

import consul
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from motor.motor_asyncio import AsyncIOMotorClient
from pydantic import BaseModel

# 导入我们的gRPC客户端
from grpc_client.crawler_client import CrawlerClient, CrawlerServiceConfig

# 加载环境变量
load_dotenv()

# 创建应用
app = FastAPI(
    title="IntelliInsight 智析",
    description="分布式爬虫与数据分析系统",
    version="1.0.0"
)

# 允许跨域
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 数据模型
class CrawlRequest(BaseModel):
    keyword: str
    post_count: int = 1
    include_comments: bool = False
    min_likes: int = 0
    comments_per_post: int = 10
    replies_per_comment: int = 5
    include_images: bool = True


# 全局变量
mongodb_client = None
crawler_client = None
consul_client = None


@app.on_event("startup")
async def startup_db_client():
    global mongodb_client, crawler_client, consul_client

    # 连接MongoDB
    mongodb_uri = os.getenv("MONGODB_URI")
    mongodb_db = os.getenv("MONGODB_DATABASE")
    mongodb_client = AsyncIOMotorClient(mongodb_uri)
    app.mongodb = mongodb_client[mongodb_db]

    # 连接Consul并发现爬虫服务
    consul_host = os.getenv("CONSUL_HOST", "localhost")
    consul_port = int(os.getenv("CONSUL_PORT", "8500"))

    try:
        consul_client = consul.Consul(host=consul_host, port=consul_port)

        # 从Consul获取爬虫服务地址
        services = consul_client.catalog.service("consul-crawler.rpc")[1]
        if services:
            service = services[0]
            crawler_host = service['ServiceAddress']
            crawler_port = service['ServicePort']
        else:
            # 开发环境默认地址
            crawler_host = os.getenv("CRAWLER_HOST", "localhost")
            crawler_port = int(os.getenv("CRAWLER_PORT", "50051"))
    except Exception as e:
        # 如果Consul连接失败，使用默认配置
        crawler_host = os.getenv("CRAWLER_HOST", "localhost")
        crawler_port = int(os.getenv("CRAWLER_PORT", "50051"))

    # 创建爬虫客户端
    config = CrawlerServiceConfig(
        host=crawler_host,
        port=crawler_port,
        timeout=int(os.getenv("GRPC_TIMEOUT", "30")),
        max_retries=int(os.getenv("GRPC_MAX_RETRIES", "3"))
    )
    crawler_client = CrawlerClient(config)


@app.on_event("shutdown")
async def shutdown_db_client():
    global mongodb_client, crawler_client
    if mongodb_client:
        mongodb_client.close()
    if crawler_client:
        crawler_client.close()


# API路由
@app.post("/api/crawl", response_model=dict)
async def start_crawl(request: CrawlRequest):
    """启动新的爬虫任务"""
    try:
        return crawler_client.start_crawl(
            keyword=request.keyword,
            post_count=request.post_count,
            include_comments=request.include_comments,
            min_likes=request.min_likes,
            comments_per_post=request.comments_per_post,
            replies_per_comment=request.replies_per_comment,
            include_images=request.include_images
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"爬虫服务调用失败: {str(e)}")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)