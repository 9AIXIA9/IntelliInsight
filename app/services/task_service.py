import logging
from datetime import datetime
from typing import Dict, Optional, Any

from app.api.models.task import TaskStatus, TaskDetail
from app.core.config import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


def map_task_status(status_code: int) -> str:
    """将整数状态码映射到字符串状态"""
    status_map = {
        0: "completed",
        1: "failed",
        2: "running",
        3: "pending",
        4: "divided"
    }
    return status_map.get(status_code, "unknown")


class TaskService:
    def __init__(self, db):
        self.db = db
        self.collection = self.db[settings.MONGODB_TASK_COLLECTION]

    async def get_tasks(self, skip: int = 0, limit: int = None,
                        status: Optional[str] = None, keyword: Optional[str] = None) -> Dict[str, Any]:
        """获取任务列表"""
        # 使用配置的默认值
        if limit is None:
            limit = settings.DEFAULT_PAGE_SIZE

        # 限制页面大小
        limit = min(limit, settings.MAX_PAGE_SIZE)

        # 构建查询条件
        query = {}
        if status:
            query["status"] = status
        if keyword:
            query["keyword"] = {"$regex": keyword, "$options": settings.SEARCH_OPTIONS}

        # 计算总数
        total = await self.collection.count_documents(query)

        # 查询数据
        cursor = self.collection.find(query).skip(skip).limit(limit).sort(
            "start_time", settings.POST_SORT_ORDER
        )

        # 转换数据格式
        tasks = []
        async for doc in cursor:
            task = {
                "task_id": str(doc["_id"]),
                "keyword": doc.get("keyword", ""),
                "site": doc.get("site", 0),
                "status": self._convert_status(doc.get("status", 3)),  # 默认为pending
                "created_at": doc.get("start_time", datetime.now()),
                "completed_at": doc.get("end_time"),
                "posts_collected": doc.get("posts_collected", 0)
            }
            tasks.append(task)

        # 返回结果
        return {
            "tasks": tasks,
            "total": total,
            "page": skip // limit + 1,
            "page_size": limit
        }

    def _convert_status(self, status_code: int) -> str:
        """将状态码转换为字符串表示"""
        status_map = {
            0: "completed",
            1: "failed",
            2: "running",
            3: "pending",
            4: "divided"
        }
        return status_map.get(status_code, "pending")

    async def get_task(self, task_id: str) -> Optional[TaskDetail]:
        """获取任务详情"""
        doc = await self.collection.find_one({"_id": task_id})
        if not doc:
            return None

        return TaskDetail(
            task_id=str(doc["_id"]),
            keyword=doc["keyword"],
            site=doc["site"],
            status=TaskStatus(map_task_status(doc["status"])),
            created_at=doc["start_time"],
            completed_at=doc.get("end_time"),
            posts_collected=doc.get("posts_collected", 0),
            post_count=doc["post_count"],
            include_comments=doc["include_comments"],
            min_likes=doc["min_likes"],
            comments_per_post=doc["comments_per_post"],
            comment_min_likes=doc["comment_min_likes"],
            include_images=doc["include_images"],
            error_message=str(doc["err"]) if doc.get("err") else None
        )
