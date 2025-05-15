from datetime import datetime
from typing import List, Optional

from bson import ObjectId
from pydantic import BaseModel, Field


class PyObjectId(ObjectId):
    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not ObjectId.is_valid(v):
            raise ValueError("无效的ObjectId")
        return ObjectId(v)

    @classmethod
    def __modify_schema__(cls, field_schema):
        field_schema.update(type="string")


class Comment(BaseModel):
    id: PyObjectId = Field(default_factory=PyObjectId, alias="_id")
    comment_id: str
    content: str
    author: str
    likes: int = 0
    comment_time: datetime

    class Config:
        json_encoders = {ObjectId: str}
        allow_population_by_field_name = True


class Post(BaseModel):
    id: PyObjectId = Field(default_factory=PyObjectId, alias="_id")
    post_id: str
    title: str
    content: str
    author: str
    likes: int = 0
    publish_time: datetime
    images: List[str] = []
    comments: List[Comment] = []
    tags: List[str] = []
    rating: float = 0.0
    location: Optional[str] = None
    category: str = ""
    task_id: str

    class Config:
        json_encoders = {ObjectId: str}
        allow_population_by_field_name = True
