import logging
import os
from contextlib import asynccontextmanager

import consul
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from motor.motor_asyncio import AsyncIOMotorClient
from pydantic import BaseModel, Field

from grpc_client.crawler_client import CrawlerClient, CrawlerServiceConfig

XIAOHONGSHU = 0  # 小红书站点的枚举值

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 加载环境变量
load_dotenv()


@asynccontextmanager
async def lifespan(app: FastAPI):
    global mongodb_client, crawler_config, consul_client

    # 连接MongoDB
    mongodb_uri = os.getenv("MONGODB_URI")
    mongodb_db = os.getenv("MONGODB_DATABASE")
    logger.info(f"连接MongoDB: {mongodb_uri}")
    mongodb_client = AsyncIOMotorClient(mongodb_uri)
    app.mongodb = mongodb_client[mongodb_db]

    # 连接Consul并发现爬虫服务
    consul_host = os.getenv("CONSUL_HOST", "localhost")
    consul_port = int(os.getenv("CONSUL_PORT", "8500"))
    consul_service_key = os.getenv("CONSUL_SERVICE_KEY", "consul-crawler.rpc")

    try:
        logger.info(f"连接Consul: {consul_host}:{consul_port}")
        consul_client = consul.Consul(host=consul_host, port=consul_port)

        # 从Consul获取爬虫服务地址
        services = consul_client.catalog.service(consul_service_key)[1]
        if services and len(services) > 0:
            service = services[0]
            crawler_host = service.get('ServiceAddress') or service.get('Address')
            crawler_port = service.get('ServicePort')
            logger.info(f"从Consul发现爬虫服务: {crawler_host}:{crawler_port}")
        else:
            # 开发环境默认地址
            crawler_host = os.getenv("CRAWLER_HOST", "localhost")
            crawler_port = int(os.getenv("CRAWLER_PORT", "50051"))
            logger.info(f"使用默认爬虫服务配置: {crawler_host}:{crawler_port}")
    except Exception as e:
        logger.error(f"Consul连接失败: {str(e)}，使用默认配置")
        crawler_host = os.getenv("CRAWLER_HOST", "localhost")
        crawler_port = int(os.getenv("CRAWLER_PORT", "50051"))

    # 创建爬虫客户端配置
    crawler_config = CrawlerServiceConfig(
        host=crawler_host,
        port=crawler_port,
        timeout=int(os.getenv("GRPC_TIMEOUT", "30")),
        max_retries=int(os.getenv("GRPC_MAX_RETRIES", "3"))
    )

    yield  # 这里是应用运行的部分

    # 关闭连接，原shutdown逻辑
    if mongodb_client:
        logger.info("关闭MongoDB连接")
        mongodb_client.close()


# 创建应用
app = FastAPI(
    title="IntelliInsight 智析",
    description="分布式爬虫与数据分析系统",
    version="1.0.1",
    lifespan=lifespan
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
    keyword: str = Field(default="南昌", description="爬取关键词，默认为南昌")
    site: int = Field(default=XIAOHONGSHU, description="爬取站点，默认为小红书")
    post_count: int = 1
    include_comments: bool = False
    min_likes: int = 0
    comments_per_post: int = 10
    replies_per_comment: int = 5
    include_images: bool = True


def get_crawler_client():
    """依赖注入获取爬虫客户端"""
    client = CrawlerClient(crawler_config)
    try:
        # 显式连接
        client.connect()

        # 验证连接是否成功
        if client._stub is None:
            logger.error("爬虫服务连接失败: stub对象为None")
            raise RuntimeError("爬虫服务连接失败")

        logger.info(f"成功连接到爬虫服务: {crawler_config.address}")
        yield client
    finally:
        client.close()


# API路由
@app.post("/api/crawl", response_model=dict)
async def start_crawl(request: CrawlRequest, client: CrawlerClient = Depends(get_crawler_client)):
    """启动新的爬虫任务"""
    try:
        logger.info(f"启动爬虫任务: {request.keyword}, 站点: {request.site}")
        # 不使用with语句，因为依赖注入已经处理了连接关闭
        return client.start_crawl(
            site=request.site,
            keyword=request.keyword,
            post_count=request.post_count,
            include_comments=request.include_comments,
            min_likes=request.min_likes,
            comments_per_post=request.comments_per_post,
            replies_per_comment=request.replies_per_comment,
            include_images=request.include_images
        )
    except Exception as e:
        logger.error(f"爬虫服务调用失败: {str(e)}")
        raise HTTPException(status_code=500, detail=f"爬虫服务调用失败: {str(e)}")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8000)
