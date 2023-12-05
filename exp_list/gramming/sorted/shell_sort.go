package sorted

import "fmt"

// 1.计算间隔
// 2.每一个间隔中的数据进行排序
// 3.缩小间隔继续排序
func shellSort(s []int) []int {
	length := len(s)
	// 计算初始的间隔
	gap := 1
	for gap < length/3 {
		gap = gap*3 + 1
	}

	for gap > 0 {
		for i := gap; i < length; i++ {
			temp := s[i]
			j := i

			// 向前比较并移动元素
			for j >= gap && s[j-gap] > temp {
				s[j] = s[j-gap]
				j -= gap
			}
			s[j] = temp
		}
		// 更新间隔
		gap = gap / 3
	}
	return s
}

func ShowShellSort() {
	fmt.Println(">希尔排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := shellSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
