package sorted

import "fmt"

func mergeSort(s []int) []int {
	if len(s) <= 1 {
		return s
	}
	mid := len(s) / 2
	left := mergeSort(s[:mid])
	right := mergeSort(s[mid:])
	return merge(left, right)
}

func merge(left, right []int) []int {
	l, r := len(left), len(right)
	res := make([]int, 0, l+r)
	i, j := 0, 0
	for i < l && j < r {
		if left[i] <= right[j] {
			res = append(res, left[i])
			i++
		} else {
			res = append(res, right[j])
			j++
		}
	}
	res = append(res, left[i:]...)
	res = append(res, right[j:]...)
	return res
}

func ShowMergeSort() {
	fmt.Println(">归并排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := mergeSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
