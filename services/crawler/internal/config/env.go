package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/logx"
	"os"
)

func LoadEnv() {
	//加载根目录的 .env（覆盖现有变量）
	if err := godotenv.Overload(); err != nil {
		logx.Severef("load .env failed,err: %v", err)
	}

	//获取并设置环境变量 APP_ENV
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
		logx.Infof("APP_ENV not set, defaulting to: %s", env)
	}

	//加载环境专属的 .env 文件（如 .env.dev）
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Overload(envFile); err != nil {
		logx.Severef("failed to load environment file %s: %v", envFile, err)
	}
}
