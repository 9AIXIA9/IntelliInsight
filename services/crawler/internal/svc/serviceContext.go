package svc

import (
	"crawler/internal/config"
	"crawler/internal/queue"
)

type ServiceContext struct {
	Config    config.Config
	TaskQueue queue.TaskQueue
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		TaskQueue: queue.NewTaskQueue(),
	}
}
