from motor.motor_asyncio import AsyncIOMotorClient

from .config import settings


class MongoDB:
    client: AsyncIOMotorClient = None
    db = None


db = MongoDB()


async def connect_to_mongo():
    db.client = AsyncIOMotorClient(settings.MONGODB_URI)
    db.db = db.client[settings.MONGODB_DB]
    print("MongoDB连接成功")


async def close_mongo_connection():
    if db.client:
        db.client.close()
        print("MongoDB连接已关闭")
