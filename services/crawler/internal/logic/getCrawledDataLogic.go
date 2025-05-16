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

// GetCrawledData 获取已爬取的数据
func (l *GetCrawledDataLogic) GetCrawledData(in *proto.DataRequest) (*proto.DataResponse, error) {
	// 检查任务ID
	if len(in.TaskId) == 0 {
		return &proto.DataResponse{
			Items:      []*proto.PostItem{},
			TotalCount: 0,
			HasMore:    false,
		}, nil
	}

	// 分页参数检查和默认值设置
	if in.Limit <= 0 {
		in.Limit = 10 // 默认每页10条数据
	}
	if in.Offset < 0 {
		in.Offset = 0
	}

	// 从任务队列获取爬取数据
	data, err := l.svcCtx.TaskQueue.GetTaskData(in.TaskId, in.Offset, in.Limit)
	if err != nil {
		l.Logger.Errorf("获取任务数据失败: %s, 错误: %v", in.TaskId, err)
		return &proto.DataResponse{
			Items:      []*proto.PostItem{},
			TotalCount: 0,
			HasMore:    false,
		}, nil
	}

	l.Logger.Infof("获取任务数据: %s, 偏移量: %d, 限制: %d, 返回条目: %d",
		in.TaskId, in.Offset, in.Limit, len(data.Items))

	return data, nil
}
