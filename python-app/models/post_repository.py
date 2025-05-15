from typing import List

from core.database import db
from models.post_model import Post


class PostRepository:
    collection = "posts"

    @classmethod
    async def find_by_task_id(cls, task_id: str, skip: int = 0, limit: int = 20) -> List[Post]:
        cursor = db.db[cls.collection].find({"task_id": task_id}).skip(skip).limit(limit)
        return [Post(**post) for post in await cursor.to_list(length=limit)]

    @classmethod
    async def count_by_task_id(cls, task_id: str) -> int:
        return await db.db[cls.collection].count_documents({"task_id": task_id})

    @classmethod
    async def save_many(cls, posts: List[dict]) -> bool:
        if not posts:
            return False
        result = await db.db[cls.collection].insert_many(posts)
        return len(result.inserted_ids) > 0
