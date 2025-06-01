from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel, Field


class Comment(BaseModel):
    id: str
    commenter: str
    content: str
    time: datetime
    like_count: int
    reply_count: int
    location: Optional[str] = None


class Post(BaseModel):
    id: str
    task_id: str
    title: str
    poster: str
    content: str
    time: datetime
    like_count: int
    comment_count: int
    collect_count: int
    location: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    image_urls: List[str] = Field(default_factory=list)
    comments: List[Comment] = Field(default_factory=list)


class PostList(BaseModel):
    posts: List[Post]
    total: int
    page: int
    page_size: int
