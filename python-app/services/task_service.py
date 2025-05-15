from datetime import datetime
from typing import List, Dict, Any

from core.database import db
from grpc_client.crawler_client import CrawlerClient


class TaskService:
    @staticmethod
    async def get_task_history(skip: int = 0, limit: int = 10) -> List[Dict[str, Any]]:
        cursor = db.db["tasks"].find().sort("created_at", -1).skip(skip).limit(limit)
        return await cursor.to_list(length=limit)

    @staticmethod
    async def create_task_record(task_id: str, keyword: str, page_count: int, category: str = "") -> str:
        task = {
            "task_id": task_id,
            "keyword": keyword,
            "page_count": page_count,
            "category": category,
            "created_at": datetime.utcnow(),
            "status": "PENDING"
        }
        result = await db.db["tasks"].insert_one(task)
        return str(result.inserted_id)

    @staticmethod
    async def update_task_status(task_id: str, status: str, progress: float = 0):
        await db.db["tasks"].update_one(
            {"task_id": task_id},
            {"$set": {"status": status, "progress": progress, "updated_at": datetime.utcnow()}}
        )
