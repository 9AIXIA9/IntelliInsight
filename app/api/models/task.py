from datetime import datetime
from enum import Enum
from typing import Optional, List

from pydantic import BaseModel


class TaskStatus(str, Enum):
    COMPLETED = "completed"
    FAILED = "failed"
    RUNNING = "running"
    PENDING = "pending"
    DIVIDED = "divided"


class TaskListItem(BaseModel):
    task_id: str
    keyword: str
    site: int
    status: TaskStatus
    created_at: datetime
    completed_at: Optional[datetime] = None
    posts_collected: int = 0


class TaskDetail(TaskListItem):
    post_count: int
    include_comments: bool
    min_likes: int
    comments_per_post: int
    comment_min_likes: int
    include_images: bool
    error_message: Optional[str] = None


class TaskList(BaseModel):
    tasks: List[TaskListItem]
    total: int
    page: int
    page_size: int
