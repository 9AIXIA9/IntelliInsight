from fastapi import APIRouter, HTTPException, Request, Depends

from app.api.models.requests import CrawlRequest
from app.api.models.response import CrawlResponse
from app.services.crawler_service import CrawlerService

router = APIRouter(prefix="/api", tags=["crawl"])


def get_crawler_service(request: Request) -> CrawlerService:
    """依赖注入：获取爬虫服务"""
    return request.app.state.crawler_service  # 使用已存在的服务实例


@router.post("/crawl", response_model=CrawlResponse)
async def start_crawl(
        request_data: CrawlRequest,
        service: CrawlerService = Depends(get_crawler_service)
):
    """启动新的爬虫任务"""
    try:
        result = await service.start_crawl(request_data.model_dump())
        return CrawlResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"爬虫服务调用失败: {str(e)}")
