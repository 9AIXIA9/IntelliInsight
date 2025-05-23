import logging
from typing import Tuple

import consul

from app.core.config import get_settings

settings = get_settings()
logger = logging.getLogger(__name__)


async def discover_service() -> Tuple[str, int]:
    """发现爬虫服务地址"""
    try:
        c = consul.Consul(host=settings.CONSUL_HOST, port=settings.CONSUL_PORT)
        _, service_info = c.kv.get(settings.CONSUL_SERVICE_KEY)

        if service_info and service_info['Value']:
            service_str = service_info['Value'].decode('utf-8')
            host, port = service_str.split(':')
            return host, int(port)
    except Exception as e:
        logger.error(f"服务发现失败: {str(e)}")

    # 返回默认配置
    return settings.CRAWLER_HOST, settings.CRAWLER_PORT
