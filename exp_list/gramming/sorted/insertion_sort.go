package sorted

import "fmt"

func insertionSort(s []int) []int {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		// 向前比较并移动元素
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
	return s
}

func ShowInsertionSort() {
	fmt.Println(">插入排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := insertionSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
