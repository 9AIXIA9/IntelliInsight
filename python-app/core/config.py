import os

from dotenv import load_dotenv
from pydantic_settings import BaseSettings

load_dotenv()


class Settings(BaseSettings):
    APP_NAME: str = "爬虫数据分析平台"
    API_PREFIX: str = "/api"
    DEBUG: bool = os.getenv("DEBUG", "False").lower() == "true"

    # MongoDB配置
    MONGODB_URI: str = os.getenv("MONGODB_URI", "mongodb://localhost:27017")
    MONGODB_DB: str = os.getenv("MONGODB_DB", "crawler")

    # gRPC服务配置
    GRPC_HOST: str = os.getenv("GRPC_HOST", "localhost")
    GRPC_PORT: int = int(os.getenv("GRPC_PORT", "8080"))


settings = Settings()
