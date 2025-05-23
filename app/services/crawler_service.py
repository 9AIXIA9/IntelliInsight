import logging
import time
from typing import Dict, Any

import grpc

from app.grpc_client.crawler_client import CrawlerClient, CrawlerServiceConfig

logger = logging.getLogger(__name__)


class CrawlerService:
    def __init__(self, config: Dict[str, Any]):
        self.client_config = CrawlerServiceConfig(
            host=config["host"],
            port=config["port"],
            timeout=config["timeout"],
            max_retries=config["max_retries"]
        )
        self.client = None
        # 初始化时就创建连接
        self._get_client()

    def _get_client(self):
        """获取或创建客户端连接"""
        if self.client is None:
            self.client = CrawlerClient(self.client_config)
            self.client.connect()
        return self.client

    async def start_crawl(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """启动爬虫任务"""
        max_attempts = 2  # 连接级别重试
        for attempt in range(max_attempts):
            try:
                client = self._get_client()
                return client.start_crawl(**request_data)
            except grpc.RpcError as e:
                logger.error(f"gRPC调用错误 (尝试 {attempt + 1}/{max_attempts}): {str(e)}")
                # 连接错误时重置连接
                if self.client:
                    try:
                        self.client.close()
                    except:
                        pass
                    self.client = None

                if attempt < max_attempts - 1:
                    time.sleep(1)  # 等待1秒后重试
            except Exception as e:
                logger.error(f"爬虫服务调用失败: {str(e)}")
                raise

        raise Exception("爬虫服务调用失败，超过最大重试次数")
