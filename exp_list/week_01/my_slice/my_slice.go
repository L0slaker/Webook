package my_slice

import "errors"

// DeleteSliceV1 实现删除操作
func DeleteSliceV1(s []any, idx int) []any {
	// 无效索引
	if len(s) <= 1 || idx < 0 || idx >= len(s) {
		return nil
	}
	res := make([]any, len(s)-1)

	// 将idx前的元素添加到新切片res
	res = append(res[:idx], s[:idx]...)
	// 将idx后的元素添加到新切片res
	res = append(res[idx:], s[idx+1:]...)

	return res
}

// DeleteSliceV2 使用比较高性能的实现：不创建新的切片，而是通过原切片的移位来实现删除
func DeleteSliceV2(s []any, idx int) []any {
	if len(s) <= 1 || idx < 0 || idx >= len(s) {
		return nil
	}

	//将idx后的元素往前移动
	copy(s[idx:], s[idx+1:])
	//缩减切片长度，丢弃最后一个元素
	s = s[:len(s)-1]
	return s
}

// DeleteSliceV3 改造为泛型方法
func DeleteSliceV3[T any](s []T, idx int) []T {
	if len(s) <= 1 || idx < 0 || idx >= len(s) {
		return nil
	}
	//将idx后的元素往前移动
	copy(s[idx:], s[idx+1:])
	//缩减切片长度，丢弃最后一个元素
	s = s[:len(s)-1]
	return s
}

// DeleteSliceV4 支持缩容
func DeleteSliceV4[T any](s []T, idx int) ([]T, T, error) {
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
