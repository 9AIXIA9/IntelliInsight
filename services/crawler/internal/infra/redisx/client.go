package redisx

import (
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// NewClient 创建Redis客户端
func NewClient(conf redis.RedisConf) (*redis.Redis, error) {
	// 创建客户端
	client, err := redis.NewRedis(conf)
	if err != nil {
		return nil, err
	}

	// 验证连接
	ok := client.Ping()
	if !ok {
		return nil, errors.New("无法连接Redis")
	}
	return client, nil
}

// MustNewClient 创建Redis客户端，连接失败时panic
func MustNewClient(conf redis.RedisConf) *redis.Redis {
	client, err := NewClient(conf)
	if err != nil {
		logx.Severef("连接Redis错误：%v", err)
	}

	return client
}
