package utils

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// GetByRound 轮询获取
// 根据id轮询获取切片中的元素，适用于负载均衡场景
func GetByRound[T any](id uint32, slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	if len(slice) == 1 {
		return slice[0]
	}
	return slice[int(id)%len(slice)]
}

// GetByRound3 轮询获取三种类型的元素
func GetByRound3[T1 any, T2 any, T3 any](id uint32, slice1 []T1, slice2 []T2, slice3 []T3) (T1, T2, T3) {
	return GetByRound(id, slice1), GetByRound(id, slice2), GetByRound(id, slice3)
}

// GetRandomly 随机获取
// 随机获取切片中的元素，适用于负载均衡场景
func GetRandomly[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	if len(slice) == 1 {
		return slice[0]
	}
	return slice[rand.IntN(len(slice))]
}

// GetRandomly3 随机获取三种类型的元素
func GetRandomly3[T1 any, T2 any, T3 any](slice1 []T1, slice2 []T2, slice3 []T3) (T1, T2, T3) {
	return GetRandomly(slice1), GetRandomly(slice2), GetRandomly(slice3)
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

func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixMicro())
}

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
