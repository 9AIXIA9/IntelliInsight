package svc

import (
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/ctrl"
	"crawler/internal/infra/mongo"
	"crawler/internal/infra/resource"
	"github.com/zeromicro/go-zero/core/stores/mon"
)

type ServiceContext struct {
	Config    *config.Config
	TaskQueue domain.TaskQueue
	Repo      domain.Repository
}

func MustNewServiceContext(c *config.Config) *ServiceContext {
	// 使用配置中的集合名称
	postModel := mon.MustNewModel(c.MongoDB.URI, c.MongoDB.Database, c.MongoDB.PostCollection)
	taskModel := mon.MustNewModel(c.MongoDB.URI, c.MongoDB.Database, c.MongoDB.TaskCollection)

	// 初始化MongoDB存储仓库
	repo := mongo.NewRepository(postModel, taskModel)

	//初始化资源池
	resourcePool := resource.MustNewResourcePool(c.Resource)

	// 使用配置的工作线程数初始化任务队列
	taskQueue := ctrl.NewTaskQueue(&c.TaskQueue, repo, resourcePool)

	return &ServiceContext{
		Config:    c,
		TaskQueue: taskQueue,
		Repo:      repo,
	}
}
