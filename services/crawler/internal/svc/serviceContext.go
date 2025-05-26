package svc

import (
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/ctrl"
	"crawler/internal/infra/mongo"
	"crawler/internal/infra/redisx"
	"crawler/internal/infra/resource"
	"github.com/zeromicro/go-zero/core/bloom"
)

type ServiceContext struct {
	Config    *config.Config
	TaskQueue domain.TaskQueue
	Repo      domain.Repository
}

func MustNewServiceContext(c *config.Config) *ServiceContext {
	// 创建MongoDB客户端
	mongoClient := mongo.MustNewClient(c.MongoDB.URI)

	// 初始化MongoDB存储仓库
	repo := mongo.NewRepository(
		mongoClient,
		c.MongoDB.Database,
		c.MongoDB.PostCollection,
		c.MongoDB.CommentCollection,
		c.MongoDB.TaskCollection,
	)

	//创建Redis客户端
	redisClient := redisx.MustNewClient(c.Redisx)

	// 初始化布隆过滤器
	filter := bloom.New(redisClient, c.BloomFilter.Key, c.BloomFilter.Bits) //并发安全

	//初始化资源池
	resourcePool := resource.MustNewResourcePool(c.Resource)

	// 使用配置的工作线程数初始化任务队列
	taskQueue := ctrl.NewTaskQueue(&c.TaskQueue, repo, filter, resourcePool)

	return &ServiceContext{
		Config:    c,
		TaskQueue: taskQueue,
		Repo:      repo,
	}
}
