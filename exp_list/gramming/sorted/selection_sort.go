package sorted

import "fmt"

func selectionSort(s []int) []int {
	length := len(s)
	for i := 0; i < length-1; i++ {
		min := i
		// 最小索引
		for j := i + 1; j < length; j++ {
			if s[j] < s[min] {
				min = j
			}
		}
		s[i], s[min] = s[min], s[i]
	}
	return s
}

func ShowSelectionSort() {
	fmt.Println(">选择排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := selectionSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
