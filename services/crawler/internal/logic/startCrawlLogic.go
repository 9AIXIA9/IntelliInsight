package logic

import (
	"context"
	"crawler/internal/infra/ctrl"

	"crawler/internal/svc"
	"crawler/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartCrawlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartCrawlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartCrawlLogic {
	return &StartCrawlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StartCrawl 启动爬虫任务
func (l *StartCrawlLogic) StartCrawl(in *proto.CrawlRequest) (*proto.CrawlResponse, error) {
	// 参数检查
	if len(in.Keyword) == 0 {
		return &proto.CrawlResponse{
			Success: false,
			Message: "搜索关键词不能为空",
		}, nil
	}

	if in.PostCount <= 0 {
		in.PostCount = 1 // 默认至少爬取1页
	}

	// 创建新任务
	task := ctrl.NewTask(in)

	// 提交任务到队列

	err := l.svcCtx.TaskQueue.AddTask(task)
	if err != nil {
		return &proto.CrawlResponse{
			TaskId:  string(task.Info.ID),
			Success: false,
			Message: err.Error(),
		}, err
	}

	l.Logger.Infof("启动新爬虫任务: %s, 关键词: %s, 页数: %d",
		task.Info.ID, in.Keyword, in.PostCount)

	return &proto.CrawlResponse{
		TaskId:  string(task.Info.ID),
		Success: true,
		Message: "任务已成功提交",
	}, nil
}
