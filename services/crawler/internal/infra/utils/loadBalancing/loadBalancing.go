package loadBalancing

import "math/rand/v2"

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
