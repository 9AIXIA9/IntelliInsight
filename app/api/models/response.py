from pydantic import BaseModel


class CrawlResponse(BaseModel):
    task_id: str
    success: bool
    message: str
