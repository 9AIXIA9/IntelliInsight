package svc

import (
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/crawler"
	"crawler/internal/infra/mongo"
	"crawler/internal/infra/resource"
	"github.com/zeromicro/go-zero/core/stores/mon"
)

type ServiceContext struct {
	Config           config.Config
	CrawlerTaskQueue domain.TaskQueue
	Repo             domain.Repository
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 使用配置中的集合名称
	postModel := mon.MustNewModel(c.MongoDB.URI, c.MongoDB.Database, c.MongoDB.PostCollection)
	taskModel := mon.MustNewModel(c.MongoDB.URI, c.MongoDB.Database, c.MongoDB.TaskCollection)

	// 初始化MongoDB存储仓库
	repo := mongo.NewRepository(postModel, taskModel)

	//初始化资源池
	pool := resource.NewResourcePool(c.Resource)

	// 使用配置的工作线程数初始化任务队列
	taskQueue := crawler.NewTaskQueue(&c.TaskQueue, repo, pool)

	return &ServiceContext{
		Config:           c,
		CrawlerTaskQueue: taskQueue,
		Repo:             repo,
	}
}
