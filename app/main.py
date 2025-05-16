import os
from typing import Optional

import consul
import grpc
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from motor.motor_asyncio import AsyncIOMotorClient
from pydantic import BaseModel

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

# 导入生成的gRPC代码
import crawler_pb2
import crawler_pb2_grpc


# 数据模型
class CrawlRequest(BaseModel):
    keyword: str
    page_count: int = 1
    include_comments: bool = False
    min_likes: int = 0
    category: Optional[str] = None


# 全局变量
mongodb_client = None
crawler_stub = None
consul_client = None


@app.on_event("startup")
async def startup_db_client():
    global mongodb_client, crawler_stub, consul_client

    # 连接MongoDB
    mongodb_uri = os.getenv("MONGODB_URI")
    mongodb_db = os.getenv("MONGODB_DATABASE")
    mongodb_client = AsyncIOMotorClient(mongodb_uri)
    app.mongodb = mongodb_client[mongodb_db]

    # 连接Consul并发现爬虫服务
    consul_host = os.getenv("CONSUL_HOST")
    consul_port = int(os.getenv("CONSUL_PORT"))
    consul_client = consul.Consul(host=consul_host, port=consul_port)

    # 从Consul获取爬虫服务地址
    services = consul_client.catalog.service("consul-crawler.rpc")[1]
    if services:
        service = services[0]
        crawler_addr = f"{service['ServiceAddress']}:{service['ServicePort']}"
    else:
        # 开发环境默认地址
        crawler_addr = "localhost:50051"

    # 创建gRPC通道和存根
    channel = grpc.insecure_channel(crawler_addr)
    crawler_stub = crawler_pb2_grpc.CrawlerServiceStub(channel)


@app.on_event("shutdown")
async def shutdown_db_client():
    global mongodb_client
    if mongodb_client:
        mongodb_client.close()


# API路由

@app.post("/api/crawl", response_model=dict)
async def start_crawl(request: CrawlRequest):
    """启动新的爬虫任务"""
    grpc_request = crawler_pb2.CrawlRequest(
        keyword=request.keyword,
        page_count=request.page_count,
        include_comments=request.include_comments,
        min_likes=request.min_likes,
        category=request.category or ""
    )

    try:
        response = crawler_stub.StartCrawl(grpc_request)
        return {
            "task_id": response.task_id,
            "success": response.success,
            "message": response.message
        }
    except grpc.RpcError as e:
        raise HTTPException(status_code=500, detail=f"gRPC调用失败: {str(e)}")


@app.get("/api/status/{task_id}")
async def get_status(task_id: str):
    """获取爬虫任务状态"""
    try:
        response = crawler_stub.GetCrawlStatus(
            crawler_pb2.StatusRequest(task_id=task_id)
        )
        return {
            "status": crawler_pb2.StatusResponse.Status.Name(response.status),
            "progress": response.progress,
            "items_collected": response.items_collected,
            "message": response.message
        }
    except grpc.RpcError as e:
        raise HTTPException(status_code=500, detail=f"gRPC调用失败: {str(e)}")


@app.get("/api/data/{task_id}")
async def get_crawl_data(
        task_id: str,
        offset: int = Query(0, ge=0),
        limit: int = Query(10, ge=1, le=100)
):
    """获取爬取的数据"""
    try:
        response = crawler_stub.GetCrawledData(
            crawler_pb2.DataRequest(
                task_id=task_id,
                offset=offset,
                limit=limit
            )
        )

        # 转换PostItem对象为字典
        items = []
        for item in response.items:
            item_dict = {
                "post_id": item.post_id,
                "title": item.title,
                "content": item.content,
                "author": item.author,
                "likes": item.likes,
                "publish_time": item.publish_time,
                "images": list(item.images),
                "tags": list(item.tags),
                "rating": item.rating,
                "location": item.location,
                "comments": []
            }

            for comment in item.comments:
                item_dict["comments"].append({
                    "comment_id": comment.comment_id,
                    "content": comment.content,
                    "author": comment.author,
                    "likes": comment.likes,
                    "comment_time": comment.comment_time
                })

            items.append(item_dict)

        return {
            "items": items,
            "total_count": response.total_count,
            "has_more": response.has_more
        }
    except grpc.RpcError as e:
        raise HTTPException(status_code=500, detail=f"gRPC调用失败: {str(e)}")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
