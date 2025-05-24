package snowflake

import (
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"github.com/bwmarrin/snowflake"
)

// GenerateID  雪花算法生成ID -> 按时间顺序增大

var (
	userIDNode *snowflake.Node
)

// MustInit 雪花算法生成ID初始化
func MustInit(startTime time.Time, machineNode int64) {
	if err := Init(startTime, machineNode); err != nil {
		logx.Severef("初始化雪花算法ID生成器错误：%v", err)
	}
}

func Init(startTime time.Time, machineNode int64) (err error) {
	snowflake.Epoch = startTime.UnixNano() / 1000000 //将起始时间转换为毫秒级时间戳
	userIDNode, err = snowflake.NewNode(machineNode)
	if err != nil {
		return err
	}

	return nil
}

// GenerateID 生成用户ID
func GenerateID() string {
	return userIDNode.Generate().String()
}
