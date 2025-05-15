from fastapi import APIRouter, Depends, HTTPException
from grpc_client.crawler_client import CrawlerClient
from pydantic import BaseModel
from services.crawler_service import get_crawler_client

router = APIRouter()


class CrawlTaskCreate(BaseModel):
    keyword: str
    page_count: int
    include_comments: bool = False
    min_likes: int = 0
    category: str = ""


@router.post("/tasks")
async def create_task(task: CrawlTaskCreate, client: CrawlerClient = Depends(get_crawler_client)):
    """创建一个新的爬虫任务"""
    try:
        response = client.start_crawl(
            keyword=task.keyword,
            page_count=task.page_count,
            include_comments=task.include_comments,
            min_likes=task.min_likes,
            category=task.category
        )
        return {"task_id": response.task_id, "message": "爬虫任务已创建"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"创建任务失败: {str(e)}")


@router.get("/tasks/{task_id}/status")
async def get_task_status(task_id: str, client: CrawlerClient = Depends(get_crawler_client)):
    """获取爬虫任务状态"""
    try:
        status = client.get_crawl_status(task_id)
        return {"task_id": task_id, "status": status.status, "progress": status.progress}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"获取任务状态失败: {str(e)}")


@router.get("/tasks/{task_id}/data")
async def get_task_data(task_id: str, offset: int = 0, limit: int = 10,
                        client: CrawlerClient = Depends(get_crawler_client)):
    """获取爬虫任务数据"""
    try:
        data = client.get_crawled_data(task_id, offset, limit)
        return {"items": [dict(item) for item in data.items], "total": data.total_count}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"获取任务数据失败: {str(e)}")
