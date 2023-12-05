package my_slice

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteSliceV1_V2_V3(t *testing.T) {
	testCases := []struct {
		name      string
		s         []any
		wantSlice []any
		index     int
		usage     func([]any, int) []any
	}{
		{
			name:      "V1",
			usage:     DeleteSliceV1,
			s:         []any{"1", "2", "3", "4", "5"},
			index:     2,
			wantSlice: []any{"1", "2", "4", "5"},
		},
		{
			name:      "V2",
			usage:     DeleteSliceV2,
			s:         []any{"1", "2", "3", "4", "5"},
			index:     2,
			wantSlice: []any{"1", "2", "4", "5"},
		},
		{
			name:      "V3",
			usage:     DeleteSliceV3[any],
			s:         []any{"1", "2", "3", "4"},
			index:     2,
			wantSlice: []any{"1", "2", "4"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println("未删除前的容量：", cap(tc.s))
			res := tc.usage(tc.s, tc.index)
			assert.Equal(t, tc.wantSlice, res)
			fmt.Println("删除后的容量：", cap(res))
		})
	}
}

// 未完成
//func TestDeleteSliceV4(t *testing.T) {
//	slice := []int{1, 2, 3, 4, 5}
//	// 删除索引为2的元素
//	result := DeleteSliceV4(slice, 2)
//	fmt.Println(result) // 输出：[1 2 4 5]
//
//	// 测试缩容功能
//	slice2 := make([]int, 16)
//	copy(slice2, []int{1, 2, 3, 4, 5})
//	fmt.Println("原始切片容量：", cap(slice2)) // 输出：原始切片容量：16
//
//	// 删除元素后触发缩容
//	DeleteSliceV4(slice2, 1)
//	DeleteSliceV4(slice2, 2)
//	result2 := DeleteSliceV4(slice2, 3)
//	fmt.Println(result2)                // 输出：[]
//	fmt.Println("新切片容量：", cap(result2)) // 输出：新切片容量：
//}
