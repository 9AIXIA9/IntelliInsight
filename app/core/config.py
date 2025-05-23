from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # 应用配置
    APP_NAME: str = "IntelliInsight 智析"
    APP_VERSION: str = "1.0.1"
    APP_DESCRIPTION: str = "分布式爬虫与数据分析系统"
    DEBUG: bool = False

    # MongoDB配置
    MONGODB_URI: str = Field(default="mongodb://localhost:27017")
    MONGODB_DATABASE: str = Field(default="crawler")
    MONGODB_POST_COLLECTION: str = Field(default="posts")
    MONGODB_TASK_COLLECTION: str = Field(default="tasks")

    # Consul配置
    CONSUL_HOST: str = Field(default="localhost")
    CONSUL_PORT: int = Field(default=8500)
    CONSUL_SERVICE_KEY: str = Field(default="consul-crawler.rpc")

    # 爬虫服务配置
    CRAWLER_HOST: str = Field(default="localhost")
    CRAWLER_PORT: int = Field(default=8999)
    GRPC_TIMEOUT: int = Field(default=30)
    GRPC_MAX_RETRIES: int = Field(default=3)

    # CORS配置
    CORS_ORIGINS: list[str] = ["http://localhost:3000"]

    class Config:
        env_file = ".env"
        case_sensitive = True


@lru_cache()
def get_settings() -> Settings:
    """获取应用配置单例"""
    return Settings()
