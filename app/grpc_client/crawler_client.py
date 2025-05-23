import logging
import time
from dataclasses import dataclass
from typing import Dict, Optional, Any, Union

import grpc

# 导入生成的protobuf代码
from .proto import crawler_pb2
from .proto import crawler_pb2_grpc

logger = logging.getLogger(__name__)


@dataclass
class CrawlerServiceConfig:
    """爬虫服务配置"""
    host: str = "localhost"
    port: int = 50051
    timeout: int = 30
    max_retries: int = 3
    retry_delay: int = 1

    @property
    def address(self) -> str:
        return f"{self.host}:{self.port}"


class CrawlerClient:
    """爬虫服务客户端封装"""

    def __init__(self, config: Optional[CrawlerServiceConfig] = None):
        """初始化爬虫客户端

        Args:
            config: 服务配置，如果为None则使用默认配置
        """
        self.config = config or CrawlerServiceConfig()
        self._channel = None
        self._stub = None

    def __enter__(self):
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def connect(self) -> None:
        """连接到爬虫gRPC服务"""
        if self._channel is None:
            try:
                logger.info(f"连接爬虫服务: {self.config.address}")
                # 使用正确的格式和选项
                address = self.config.address.replace('localhost', '127.0.0.1')
                options = [
                    ('grpc.so_reuseport', 0),
                    ('grpc.use_local_subchannel_pool', 1),
                    ('grpc.keepalive_time_ms', 30000),
                    ('grpc.keepalive_timeout_ms', 10000),
                    ('grpc.keepalive_permit_without_calls', 1)
                ]
                self._channel = grpc.insecure_channel(address)
                self._stub = crawler_pb2_grpc.CrawlerServiceStub(self._channel)
                # 检查通道状态
                try:
                    state = self._channel._channel.check_connectivity_state(True)
                    logger.info(f"爬虫服务连接状态: {state}")
                    logger.info("爬虫服务连接成功")
                except Exception as e:
                    logger.warning(f"连接状态检查失败: {e}")
            except Exception as e:
                logger.error(f"连接爬虫服务失败: {str(e)}")
                self._channel = None
                self._stub = None
                raise

    def close(self) -> None:
        """关闭连接"""
        if self._channel is not None:
            logger.info("关闭爬虫服务连接")
            self._channel.close()
            self._channel = None
            self._stub = None

    def _execute_with_retry(self, func, *args, **kwargs) -> Any:
        """执行RPC调用并处理重试

        Args:
            func: 要调用的RPC方法
            *args: 位置参数
            **kwargs: 关键字参数

        Returns:
            RPC调用结果

        Raises:
            grpc.RpcError: 如果所有重试都失败
        """
        if self._stub is None:
            self.connect()

        last_exception = None
        for attempt in range(self.config.max_retries):
            try:
                return func(*args, **kwargs)

            except grpc.RpcError as e:
                last_exception = e
                logger.warning(f"RPC调用失败 (尝试 {attempt + 1}/{self.config.max_retries}): {str(e)}")

                # 连接错误时重新连接
                # 使用安全的方式检查状态码
                status_code = e.code() if hasattr(e, 'code') else None
                if status_code in (grpc.StatusCode.UNAVAILABLE, grpc.StatusCode.DEADLINE_EXCEEDED):
                    self.close()
                    self.connect()

                if attempt < self.config.max_retries - 1:
                    time.sleep(self.config.retry_delay)

        if last_exception:
            raise last_exception

    def start_crawl(self,
                    keyword: str,
                    site: int = 0,  # 默认使用XIAOHONGSHU(0)
                    post_count: int = 10,
                    include_comments: bool = True,
                    min_likes: int = 0,
                    comments_per_post: int = 10,
                    replies_per_comment: int = 5,
                    include_images: bool = True) -> Dict[str, Union[str, bool]]:
        """启动爬虫任务

        Args:
            keyword: 搜索关键词
            site: 爬取站点，默认为小红书(0)
            post_count: 爬取帖子数量
            include_comments: 是否包含评论
            min_likes: 最少点赞数筛选
            comments_per_post: 每个帖子爬取的评论数量
            replies_per_comment: 每条评论爬取的回复数量
            include_images: 是否爬取图片URL

        Returns:
            包含任务ID、成功状态和消息的字典
        """
        request = crawler_pb2.CrawlRequest(
            site=site,
            keyword=keyword,
            post_count=post_count,
            include_comments=include_comments,
            min_likes=min_likes,
            comments_per_post=comments_per_post,
            replies_per_comment=replies_per_comment,
            include_images=include_images
        )

        # 发起RPC调用并获取响应
        response = self._execute_with_retry(self._stub.StartCrawl, request, timeout=self.config.timeout)

        # 将gRPC响应转换为字典
        return {
            'task_id': response.task_id,
            'success': response.success,
            'message': response.message
        }
