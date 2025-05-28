package loadBalancing

import (
	"errors"
	"math/rand/v2"
)

var ErrEmptySlice = errors.New("切片为空")

// -------------- 随机选择策略 --------------

// GetIndexRandomly 随机获取单个切片的索引
func GetIndexRandomly[T any](slice []T) (int, error) {
	if len(slice) == 0 {
		return 0, ErrEmptySlice
	}
	if len(slice) == 1 {
		return 0, nil
	}
	return rand.IntN(len(slice)), nil
}

// GetRandomly 随机获取单个切片的元素
func GetRandomly[T any](slice []T) (T, error) {
	var zero T
	idx, err := GetIndexRandomly(slice)
	if err != nil {
		return zero, err
	}
	return slice[idx], nil
}

// GetIndexesRandomly3 随机获取三个切片的索引
func GetIndexesRandomly3[T1, T2, T3 any](slice1 []T1, slice2 []T2, slice3 []T3) (int, int, int, error) {
	idx1, err := GetIndexRandomly(slice1)
	if err != nil {
		return 0, 0, 0, err
	}

	idx2, err := GetIndexRandomly(slice2)
	if err != nil {
		return 0, 0, 0, err
	}

	idx3, err := GetIndexRandomly(slice3)
	if err != nil {
		return 0, 0, 0, err
	}

	return idx1, idx2, idx3, nil
}

// GetRandomly3 随机获取三个切片的元素
func GetRandomly3[T1, T2, T3 any](slice1 []T1, slice2 []T2, slice3 []T3) (T1, T2, T3, error) {
	var zero1 T1
	var zero2 T2
	var zero3 T3

	if len(slice1) == 0 || len(slice2) == 0 || len(slice3) == 0 {
		return zero1, zero2, zero3, ErrEmptySlice
	}

	idx1, idx2, idx3, err := GetIndexesRandomly3(slice1, slice2, slice3)
	if err != nil {
		return zero1, zero2, zero3, err
	}

	return slice1[idx1], slice2[idx2], slice3[idx3], nil
}

// -------------- 轮询选择策略 --------------

// GetIndexByRound 根据ID轮询获取单个切片的索引
func GetIndexByRound[T any](id uint32, slice []T) (int, error) {
	if len(slice) == 0 {
		return 0, ErrEmptySlice
	}
	if len(slice) == 1 {
		return 0, nil
	}
	return int(id) % len(slice), nil
}

// GetByRound 根据ID轮询获取单个切片的元素
func GetByRound[T any](id uint32, slice []T) (T, error) {
	var zero T
	idx, err := GetIndexByRound(id, slice)
	if err != nil {
		return zero, err
	}
	return slice[idx], nil
}

// GetIndexesByRound3 根据ID轮询获取三个切片的索引
func GetIndexesByRound3[T1, T2, T3 any](id uint32, slice1 []T1, slice2 []T2, slice3 []T3) (int, int, int, error) {
	idx1, err := GetIndexByRound(id, slice1)
	if err != nil {
		return 0, 0, 0, err
	}

	idx2, err := GetIndexByRound(id, slice2)
	if err != nil {
		return 0, 0, 0, err
	}

	idx3, err := GetIndexByRound(id, slice3)
	if err != nil {
		return 0, 0, 0, err
	}

	return idx1, idx2, idx3, nil
}

// GetByRound3 根据ID轮询获取三个切片的元素
func GetByRound3[T1, T2, T3 any](id uint32, slice1 []T1, slice2 []T2, slice3 []T3) (T1, T2, T3, error) {
	var zero1 T1
	var zero2 T2
	var zero3 T3

	if len(slice1) == 0 || len(slice2) == 0 || len(slice3) == 0 {
		return zero1, zero2, zero3, ErrEmptySlice
	}

	idx1, err := GetIndexByRound(id, slice1)
	if err != nil {
		return zero1, zero2, zero3, err
	}

	idx2, err := GetIndexByRound(id, slice2)
	if err != nil {
		return zero1, zero2, zero3, err
	}

	idx3, err := GetIndexByRound(id, slice3)
	if err != nil {
		return zero1, zero2, zero3, err
	}

	return slice1[idx1], slice2[idx2], slice3[idx3], nil
}
