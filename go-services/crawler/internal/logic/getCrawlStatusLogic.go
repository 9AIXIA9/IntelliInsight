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

// 获取爬虫任务状态
func (l *GetCrawlStatusLogic) GetCrawlStatus(in *proto.StatusRequest) (*proto.StatusResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.StatusResponse{}, nil
}
