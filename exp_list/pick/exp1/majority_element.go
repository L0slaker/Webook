package main

import (
	"fmt"
	"sort"
)

// ShowMajorityElement 多数元素
func ShowMajorityElement() {
	nums1 := []int{3, 2, 3}
	nums2 := []int{2, 2, 1, 1, 1, 2, 2}
	v1 := majorityElement(nums1)
	v2 := majorityElement(nums2)
	v3 := majorityElementV2(nums1)
	v4 := majorityElementV2(nums2)
	fmt.Printf("%d 的多数元素为：%d \n", nums1, v1)
	fmt.Printf("%d 的多数元素为：%d \n", nums2, v2)
	fmt.Printf("%d 的多数元素为：%d \n", nums1, v3)
	fmt.Printf("%d 的多数元素为：%d \n", nums2, v4)
}

func majorityElement(nums []int) int {
	sort.Ints(nums)
	if len(nums) < 3 {
		return nums[0]
	}
	if nums[len(nums)/2] == nums[len(nums)/2+1] {
		return nums[len(nums)/2]
	}
	return nums[0]
}

func majorityElementV2(nums []int) int {
	counts := make(map[int]int)

	for _, num := range nums {
		counts[num]++
		// 如果某个元素的出现次数超过 n/2，则返回该元素
		if counts[num] > len(nums)/2 {
			return num
		}
	}
	return nums[0]
}
