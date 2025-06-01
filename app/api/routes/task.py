from fastapi import APIRouter, HTTPException, Request, Depends, Query

from app.api.models.task import TaskList, TaskDetail

router = APIRouter(prefix="/api/tasks", tags=["tasks"])


def get_task_service(request: Request):
    """依赖注入：获取任务服务"""
    from app.services.task_service import TaskService
    return TaskService(request.app.state.mongodb)


@router.get("", response_model=TaskList)
async def list_tasks(
        skip: int = Query(0, ge=0),
        limit: int = Query(20, ge=1, le=100),
        status: str = Query(None),
        keyword: str = Query(None),
        service=Depends(get_task_service)
):
    """获取任务列表"""
    result = await service.get_tasks(skip, limit, status, keyword)
    return result


@router.get("/{task_id}", response_model=TaskDetail)
async def get_task(task_id: str, service=Depends(get_task_service)):
    """获取任务详情"""
    task = await service.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return task
