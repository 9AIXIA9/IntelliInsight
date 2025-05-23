from motor.motor_asyncio import AsyncIOMotorClient

from app.core.config import get_settings

settings = get_settings()
_mongodb_client = None


async def connect_to_mongodb():
    """连接到MongoDB数据库"""
    global _mongodb_client
    if _mongodb_client is None:
        _mongodb_client = AsyncIOMotorClient(settings.MONGODB_URI)
    return _mongodb_client[settings.MONGODB_DATABASE]


async def close_mongodb_connection():
    """关闭MongoDB连接"""
    global _mongodb_client
    if _mongodb_client is not None:  # 检查是否为None
        _mongodb_client.close()
        _mongodb_client = None  # 关闭后重置为None
