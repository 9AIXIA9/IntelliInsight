import uvicorn
from fastapi import FastAPI

from api.routes import router as api_router
from core.config import settings
from core.database import connect_to_mongo, close_mongo_connection

app = FastAPI(
    title=settings.APP_NAME,
    debug=settings.DEBUG
)

# 启动和关闭事件
app.add_event_handler("startup", connect_to_mongo)
app.add_event_handler("shutdown", close_mongo_connection)

# 注册路由
app.include_router(api_router, prefix=settings.API_PREFIX)


# 根路由
@app.get("/")
async def root():
    return {"message": "爬虫系统API服务已启动"}


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=settings.DEBUG)
