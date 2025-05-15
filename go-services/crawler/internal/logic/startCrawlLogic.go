package logic

import (
	"context"

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

	if in.PageCount <= 0 {
		in.PageCount = 1 // 默认至少爬取1页
	}

	// 提交任务到队列
	taskID := l.svcCtx.TaskQueue.AddTask(in)

	l.Logger.Infof("启动新爬虫任务: %s, 关键词: %s, 页数: %d",
		taskID, in.Keyword, in.PageCount)

	return &proto.CrawlResponse{
		TaskId:  taskID,
		Success: true,
		Message: "任务已成功提交",
	}, nil
}
