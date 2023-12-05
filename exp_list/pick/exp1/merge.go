package main

import (
	"fmt"
	"sort"
)

func ShowMarge() {
	nums1 := []int{1, 2, 3, 0, 0, 0}
	m, n := 3, 3
	nums2 := []int{2, 5, 6}
	mergeV1(nums1, m, nums2, n)
	mergeV2(nums1, m, nums2, n)
}

func mergeV1(nums1 []int, m int, nums2 []int, n int) {
	for _, v := range nums2 {
		nums1 = append(nums1, v)
	}
	//fmt.Println("after range nums2, nums1 = ", nums1)
	length := m + n
	//fmt.Println("length = ", length)
	nLength := len(nums1)
	//fmt.Println("nLength = ", nLength)
	for i := 0; i < len(nums1)-1; i++ {
		for j := 0; j < len(nums1)-i-1; j++ {
			if nums1[j] > nums1[j+1] {
				nums1[j], nums1[j+1] = nums1[j+1], nums1[j]
			}
		}
	}
	//fmt.Println("after sorter,nums1 = ", nums1)
	nums1 = nums1[nLength-length:]
	fmt.Println("V1:", nums1)
}

func mergeV2(nums1 []int, m int, nums2 []int, _ int) {
	copy(nums1[m:], nums2)
	sort.Ints(nums1)
	fmt.Println("V2:", nums1)
}
