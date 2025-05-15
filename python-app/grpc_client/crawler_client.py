import grpc
from .crawler_pb2 import (
    CrawlRequest, StatusRequest, DataRequest
)
from .crawler_pb2_grpc import CrawlerServiceStub


class CrawlerClient:
    """Python风格的爬虫客户端"""

    def __init__(self, host="localhost", port=8080):
        """初始化gRPC连接"""
        self.channel = grpc.insecure_channel(f"{host}:{port}")
        self.client = CrawlerServiceStub(self.channel)

    def start_crawl(self, keyword, page_count, include_comments=False, min_likes=0, category=""):
        """启动爬虫任务"""
        request = CrawlRequest(
            keyword=keyword,
            page_count=page_count,
            include_comments=include_comments,
            min_likes=min_likes,
            category=category
        )
        return self.client.StartCrawl(request)

    def get_crawl_status(self, task_id):
        """获取爬虫状态"""
        request = StatusRequest(task_id=task_id)
        return self.client.GetCrawlStatus(request)

    def get_crawled_data(self, task_id, offset=0, limit=10):
        """获取已爬取的数据"""
        request = DataRequest(task_id=task_id, offset=offset, limit=limit)
        return self.client.GetCrawledData(request)

    def close(self):
        """关闭gRPC连接"""
        self.channel.close()