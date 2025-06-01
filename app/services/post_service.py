import logging
from datetime import datetime
from typing import Dict, Any, Optional, List

from app.api.models.post import Post, Comment
from app.core.config import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


class PostService:
    def __init__(self, db):
        self.db = db
        self.collection = self.db[settings.MONGODB_POST_COLLECTION]
        self.comment_collection = self.db[settings.MONGODB_COMMENT_COLLECTION]

    async def get_posts(self, skip: int = 0, limit: int = None,
                        task_id: Optional[str] = None, keyword: Optional[str] = None,
                        tag: Optional[str] = None, min_likes: Optional[int] = None) -> Dict[str, Any]:
        """获取帖子列表"""
        # 使用配置的默认值
        if limit is None:
            limit = settings.DEFAULT_PAGE_SIZE

        # 限制页面大小
        limit = min(limit, settings.MAX_PAGE_SIZE)

        # 构建查询条件
        query = {}
        if task_id:
            query["task_id"] = task_id
        if keyword:
            query["$or"] = [
                {"title": {"$regex": keyword, "$options": settings.SEARCH_OPTIONS}},
                {"content": {"$regex": keyword, "$options": settings.SEARCH_OPTIONS}}
            ]
        if tag:
            query["tags"] = {"$in": [tag]}
        if min_likes is not None:
            query["like_count"] = {"$gte": min_likes}

        # 计算总数
        total = await self.collection.count_documents(query)

        # 查询数据
        cursor = self.collection.find(query).skip(skip).limit(limit).sort(
            settings.POST_SORT_FIELD, settings.POST_SORT_ORDER
        )

        # 转换数据格式
        posts = []
        post_ids = []

        async for doc in cursor:
            # 使用get方法获取内容，提供默认值避免KeyError
            post = Post(
                id=str(doc["_id"]),
                task_id=doc["task_id"],
                title=doc.get("title", ""),
                poster=doc.get("poster", ""),
                content=doc.get("content", ""),  # 添加默认值避免KeyError
                time=doc.get("time", datetime.now()),
                like_count=doc.get("like_count", 0),
                comment_count=doc.get("comment_count", 0),
                collect_count=doc.get("collect_count", 0),
                location=doc.get("location"),
                tags=doc.get("tags", []) or [],
                image_urls=doc.get("image_urls", []) or [],
                comments=[]
            )
            posts.append(post)
            post_ids.append(doc["_id"])

        # 如果有帖子，批量查询评论
        if post_ids:
            # 确保post_ids是字符串格式
            string_post_ids = [str(pid) for pid in post_ids]

            # 为每个帖子加载评论
            await self._load_comments_for_posts(posts, string_post_ids)

        # 返回结果
        return {
            "posts": posts,
            "total": total,
            "page": skip // limit + 1,
            "page_size": limit
        }

    async def _load_comments_for_posts(self, posts: List[Post], post_ids: List[str]) -> None:
        """为帖子加载评论"""
        # 创建帖子ID到帖子对象的映射
        post_map = {post.id: post for post in posts}

        # 构建评论查询
        comment_query = {"post_id": {"$in": post_ids}}

        # 查询评论
        cursor = self.comment_collection.find(comment_query).sort(
            settings.COMMENT_SORT_FIELD, settings.COMMENT_SORT_ORDER
        )

        # 为每个帖子添加评论
        async for doc in cursor:
            post_id = doc.get("post_id")
            if post_id in post_map:
                # 只加载每个帖子的前N条评论
                if len(post_map[post_id].comments) < settings.COMMENTS_PER_POST:
                    comment = Comment(
                        id=str(doc["_id"]),
                        commenter=doc.get("commenter", ""),
                        content=doc.get("content", ""),
                        time=doc.get("time", datetime.now()),
                        like_count=doc.get("like_count", 0),
                        reply_count=doc.get("reply_count", 0),
                        location=doc.get("location")
                    )
                    post_map[post_id].comments.append(comment)
