package logic

import (
	"context"

	"crawler/internal/svc"
	"crawler/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCrawledDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCrawledDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCrawledDataLogic {
	return &GetCrawledDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取已爬取的数据
func (l *GetCrawledDataLogic) GetCrawledData(in *proto.DataRequest) (*proto.DataResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.DataResponse{}, nil
}
