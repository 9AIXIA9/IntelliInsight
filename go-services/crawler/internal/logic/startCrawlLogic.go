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

// 启动爬虫任务
func (l *StartCrawlLogic) StartCrawl(in *proto.CrawlRequest) (*proto.CrawlResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.CrawlResponse{}, nil
}
