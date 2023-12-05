package sorted

import "fmt"

func quickSort(s []int) []int {
	n := len(s)
	if n <= 1 {
		return s
	}
	// 基准元素选择最后一个元素
	pivot := s[n-1]
	left, right := []int{}, []int{}
	for i := 0; i < n-1; i++ {
		if s[i] <= pivot {
			left = append(left, s[i])
		} else {
			right = append(right, s[i])
		}
	}
	sortedLeft := quickSort(left)
	sortedRight := quickSort(right)
	res := append(append(sortedLeft, pivot), sortedRight...)
	return res
}

func ShowQuickSort() {
	fmt.Println(">快速排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := quickSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
