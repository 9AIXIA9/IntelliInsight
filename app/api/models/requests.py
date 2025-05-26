from pydantic import BaseModel, Field


class CrawlRequest(BaseModel):
    keyword: str = Field(default="南昌", description="爬取关键词，默认为南昌")
    site: int = Field(default=0, description="爬取站点，默认为小红书")
    post_count: int = Field(default=10, description="爬取帖子数量")
    include_comments: bool = Field(default=True, description="是否包含评论")
    min_likes: int = Field(default=0, description="最少点赞数")
    comments_per_post: int = Field(default=10, description="每帖评论数")
    comment_min_likes: int = Field(default=5, description="评论最少点赞数")  # 修改字段名和描述
    include_images: bool = Field(default=True, description="是否包含图片")