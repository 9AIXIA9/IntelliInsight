from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CrawlRequest(_message.Message):
    __slots__ = ("keyword", "page_count", "include_comments", "min_likes", "category")
    KEYWORD_FIELD_NUMBER: _ClassVar[int]
    PAGE_COUNT_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_COMMENTS_FIELD_NUMBER: _ClassVar[int]
    MIN_LIKES_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    keyword: str
    page_count: int
    include_comments: bool
    min_likes: int
    category: str
    def __init__(self, keyword: _Optional[str] = ..., page_count: _Optional[int] = ..., include_comments: bool = ..., min_likes: _Optional[int] = ..., category: _Optional[str] = ...) -> None: ...

class CrawlResponse(_message.Message):
    __slots__ = ("task_id", "success", "message")
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    task_id: str
    success: bool
    message: str
    def __init__(self, task_id: _Optional[str] = ..., success: bool = ..., message: _Optional[str] = ...) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ("task_id",)
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    task_id: str
    def __init__(self, task_id: _Optional[str] = ...) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("status", "progress", "items_collected", "message")
    class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        PENDING: _ClassVar[StatusResponse.Status]
        RUNNING: _ClassVar[StatusResponse.Status]
        COMPLETED: _ClassVar[StatusResponse.Status]
        FAILED: _ClassVar[StatusResponse.Status]
    PENDING: StatusResponse.Status
    RUNNING: StatusResponse.Status
    COMPLETED: StatusResponse.Status
    FAILED: StatusResponse.Status
    STATUS_FIELD_NUMBER: _ClassVar[int]
    PROGRESS_FIELD_NUMBER: _ClassVar[int]
    ITEMS_COLLECTED_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    status: StatusResponse.Status
    progress: float
    items_collected: int
    message: str
    def __init__(self, status: _Optional[_Union[StatusResponse.Status, str]] = ..., progress: _Optional[float] = ..., items_collected: _Optional[int] = ..., message: _Optional[str] = ...) -> None: ...

class DataRequest(_message.Message):
    __slots__ = ("task_id", "offset", "limit")
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    task_id: str
    offset: int
    limit: int
    def __init__(self, task_id: _Optional[str] = ..., offset: _Optional[int] = ..., limit: _Optional[int] = ...) -> None: ...

class DataResponse(_message.Message):
    __slots__ = ("items", "total_count", "has_more")
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_COUNT_FIELD_NUMBER: _ClassVar[int]
    HAS_MORE_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[PostItem]
    total_count: int
    has_more: bool
    def __init__(self, items: _Optional[_Iterable[_Union[PostItem, _Mapping]]] = ..., total_count: _Optional[int] = ..., has_more: bool = ...) -> None: ...

class PostItem(_message.Message):
    __slots__ = ("post_id", "title", "content", "author", "likes", "publish_time", "images", "comments", "tags", "rating", "location")
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_FIELD_NUMBER: _ClassVar[int]
    LIKES_FIELD_NUMBER: _ClassVar[int]
    PUBLISH_TIME_FIELD_NUMBER: _ClassVar[int]
    IMAGES_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    RATING_FIELD_NUMBER: _ClassVar[int]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    title: str
    content: str
    author: str
    likes: int
    publish_time: int
    images: _containers.RepeatedScalarFieldContainer[str]
    comments: _containers.RepeatedCompositeFieldContainer[Comment]
    tags: _containers.RepeatedScalarFieldContainer[str]
    rating: float
    location: str
    def __init__(self, post_id: _Optional[str] = ..., title: _Optional[str] = ..., content: _Optional[str] = ..., author: _Optional[str] = ..., likes: _Optional[int] = ..., publish_time: _Optional[int] = ..., images: _Optional[_Iterable[str]] = ..., comments: _Optional[_Iterable[_Union[Comment, _Mapping]]] = ..., tags: _Optional[_Iterable[str]] = ..., rating: _Optional[float] = ..., location: _Optional[str] = ...) -> None: ...

class Comment(_message.Message):
    __slots__ = ("comment_id", "content", "author", "likes", "comment_time")
    COMMENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_FIELD_NUMBER: _ClassVar[int]
    LIKES_FIELD_NUMBER: _ClassVar[int]
    COMMENT_TIME_FIELD_NUMBER: _ClassVar[int]
    comment_id: str
    content: str
    author: str
    likes: int
    comment_time: int
    def __init__(self, comment_id: _Optional[str] = ..., content: _Optional[str] = ..., author: _Optional[str] = ..., likes: _Optional[int] = ..., comment_time: _Optional[int] = ...) -> None: ...
