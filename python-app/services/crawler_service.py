from core.config import settings
from grpc_client.crawler_client import CrawlerClient


def get_crawler_client():
    client = CrawlerClient(host=settings.GRPC_HOST, port=settings.GRPC_PORT)
    try:
        yield client
    finally:
        client.close()
