package generalization_tool

import (
	"errors"
)

// Add 添加
func Add[T any](s []T, idx int, elem T) error {
	if idx < 0 || idx >= len(s) {
		return errors.New("index out of range")
	}
	_ = append(s, elem)
	copy(s[idx+1:], s[idx:])
	s[idx] = elem
	return nil
}

// Delete 删除
func Delete[T any](s []T, idx int) ([]T, T, error) {
	// 新建一个切片将原切片装进去,不指定容量
	if idx < 0 || idx >= len(s) {
		var zero T
		return nil, zero, errors.New("index out of range")
	}
	res := s[idx]

	//将idx后的元素往前移动
	for i := idx; i < len(s)-1; i++ {
		s[i] = s[i+1]
	}

	//缩减切片长度，丢弃最后一个元素
	s = s[:len(s)-1]

	// 缩容: 长度小于容量的1/4，将容量缩为原来的一半，这里设置256为缩容的阈值
	if cap(s) > 256 {
		if 4*len(s) <= cap(s) {
			newCap := cap(s) / 2
			newSlice := make([]T, len(s), newCap)
			newSlice = append(newSlice, s...)
			return newSlice, res, nil
		}
	}

	return s, res, nil
}

// Search 查找
func Search[T any](s []T, idx int) (T, error) {
	if idx < 0 || idx >= len(s) {
		var zero T
		return zero, errors.New("index out of range")
	}
	return s[idx], nil
}

// InterSet 求交集
func InterSet[T comparable](s1, s2 []T) []T {
	m := make(map[T]struct{})
	res := make([]T, 0, len(s1))

	for _, v := range s1 {
		m[v] = struct{}{}
	}

	for _, v := range s2 {
		if _, ok := m[v]; ok {
			delete(m, v)
			res = append(res, v)
		}
	}

	return res
}

// UnionSet 求并集
func UnionSet[T comparable](s1, s2 []T) []T {
	res := make([]T, 0, len(s1)+len(s2))
	m := make(map[T]struct{})

	for _, v := range s1 {
		m[v] = struct{}{}
		res = append(res, v)
	}

	for _, v := range s2 {
		if _, ok := m[v]; !ok {
			res = append(res, v)
		}
	}

	return res
}

// DiffSet 求差集
func DiffSet[T comparable](s1, s2 []T) []T {
	res := make([]T, 0, len(s1))
	m := make(map[T]struct{})

	for _, v := range s1 {
		m[v] = struct{}{}
	}

	for _, v := range s2 {
		delete(m, v)
	}

	for v := range m {
		res = append(res, v)
	}

	//顺序？
	return res
}

//// MapReduce map reduce API
//func MapReduce[T any](s []T, idx int) T {
//
//}
