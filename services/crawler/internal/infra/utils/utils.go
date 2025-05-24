package utils

import (
	"math/rand/v2"
	"time"
)

// DelayRandomly 模拟人类行为的随机延迟
// base: 基础延迟毫秒数
// 通过组合正态分布和偶发长暂停来模拟人类操作的不规则性
func DelayRandomly(base int) {
	// 基础延迟
	baseDelay := base

	// 正态分布随机因子 (模拟人类行为的集中趋势)
	// 生成[-1, 1]范围内的随机数，大多数集中在0附近
	r1 := rand.Float64()*2 - 1
	r2 := rand.Float64()*2 - 1
	// Box-Muller变换近似正态分布
	gaussian := 0.5 * (r1*r1 + r2*r2)
	// 将正态分布值映射到[0.5, 1.5]范围，使大多数值接近1
	factor := 1.0
	if gaussian <= 1.0 {
		factor = 0.5 + gaussian/2.0
	}

	// 偶发性长暂停(模拟人类思考或分心)
	if rand.Float64() < 0.1 { // 10%概率出现额外延迟
		factor += rand.Float64() * 2.0 // 最多增加基础延迟的2倍
	}

	sleepTime := time.Duration(float64(baseDelay) * factor)
	time.Sleep(sleepTime * time.Millisecond)
}

// Contains 判断切片中是否存在该值
func Contains[T comparable](set []T, element T) bool {
	for _, t := range set {
		if t == element {
			return true
		}
	}
	return false
}
