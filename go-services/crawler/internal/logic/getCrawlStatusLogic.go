package logic

import (
	"context"

	"crawler/internal/svc"
	"crawler/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCrawlStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCrawlStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCrawlStatusLogic {
	return &GetCrawlStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCrawlStatus 获取爬虫任务状态
func (l *GetCrawlStatusLogic) GetCrawlStatus(in *proto.StatusRequest) (*proto.StatusResponse, error) {
	// 检查任务ID
	if len(in.TaskId) == 0 {
		return &proto.StatusResponse{
			Status:  proto.StatusResponse_FAILED,
			Message: "无效的任务ID",
		}, nil
	}

	// 从任务队列获取任务状态
	status, err := l.svcCtx.TaskQueue.GetTaskStatus(in.TaskId)
	if err != nil {
		l.Logger.Errorf("获取任务状态失败: %s, 错误: %v", in.TaskId, err)
		return &proto.StatusResponse{
			Status:  proto.StatusResponse_FAILED,
			Message: err.Error(),
		}, nil
	}

	l.Logger.Infof("获取任务状态: %s, 状态: %v, 进度: %.2f%%",
		in.TaskId, status.Status, status.Progress)

	return status, nil
}
