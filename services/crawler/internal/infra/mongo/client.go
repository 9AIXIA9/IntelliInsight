package mongo

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewClient 创建MongoDB客户端
func NewClient(ctx context.Context, uri string) (*mongo.Client, error) {
	// 设置连接选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	// 创建客户端
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// 验证连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// MustNewClient 创建MongoDB客户端，连接失败时panic
func MustNewClient(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewClient(ctx, uri)
	if err != nil {
		logx.Severef("连接MongoDB错误：%v", err)
	}

	return client
}
