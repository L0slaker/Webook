package sorted

import "fmt"

func bubbleSort(s []int) []int {
	n := len(s)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if s[j] > s[j+1] {
				s[j], s[j+1] = s[j+1], s[j]
			}
		}
	}
	return s
}

func ShowBubbleSort() {
	fmt.Println(">冒泡排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := bubbleSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
