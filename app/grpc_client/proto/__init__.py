import importlib.util
import os
import sys

# 获取当前目录的绝对路径
current_dir = os.path.dirname(os.path.abspath(__file__))

# 将当前目录添加到Python搜索路径
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

# 动态导入模块
try:
    # 只导入实际存在的类
    from crawler_pb2 import CrawlRequest, CrawlResponse, Site
    from crawler_pb2 import DESCRIPTOR, _CRAWLREQUEST, _CRAWLRESPONSE, _CRAWLERSERVICE, _SITE
    from crawler_pb2_grpc import CrawlerServiceStub
except ImportError:
    # 如果直接导入失败，尝试更复杂的方法
    pb2_path = os.path.join(current_dir, "crawler_pb2.py")
    pb2_grpc_path = os.path.join(current_dir, "crawler_pb2_grpc.py")

    # 动态加载模块
    spec1 = importlib.util.spec_from_file_location("crawler_pb2", pb2_path)
    pb2 = importlib.util.module_from_spec(spec1)
    spec1.loader.exec_module(pb2)

    # 修复crawler_pb2_grpc的搜索路径
    sys.modules["crawler_pb2"] = pb2

    # 然后加载grpc模块
    spec2 = importlib.util.spec_from_file_location("crawler_pb2_grpc", pb2_grpc_path)
    pb2_grpc = importlib.util.module_from_spec(spec2)
    spec2.loader.exec_module(pb2_grpc)

    # 从动态加载的模块中导出，只导入实际存在的类
    CrawlRequest = pb2.CrawlRequest
    CrawlResponse = pb2.CrawlResponse
    Site = pb2.Site
    CrawlerServiceStub = pb2_grpc.CrawlerServiceStub

# 导出类以供使用
__all__ = ['CrawlRequest', 'CrawlResponse', 'Site', 'CrawlerServiceStub']
