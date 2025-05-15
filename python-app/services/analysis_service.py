# services/analysis_service.py
from typing import Dict, List, Any

from core.database import db


class AnalysisService:
    @staticmethod
    async def get_category_statistics(task_id: str) -> List[Dict[str, Any]]:
        pipeline = [
            {"$match": {"task_id": task_id}},
            {"$group": {"_id": "$category", "count": {"$sum": 1}, "avg_likes": {"$avg": "$likes"}}},
            {"$sort": {"count": -1}}
        ]
        return await db.db["posts"].aggregate(pipeline).to_list(None)

    @staticmethod
    async def get_popular_tags(task_id: str, limit: int = 10) -> List[Dict[str, Any]]:
        pipeline = [
            {"$match": {"task_id": task_id}},
            {"$unwind": "$tags"},
            {"$group": {"_id": "$tags", "count": {"$sum": 1}}},
            {"$sort": {"count": -1}},
            {"$limit": limit}
        ]
        return await db.db["posts"].aggregate(pipeline).to_list(None)
