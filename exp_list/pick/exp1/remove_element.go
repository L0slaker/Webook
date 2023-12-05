package main

import "fmt"

func ShowRemoveElement() {
	nums := []int{3, 2, 2, 3}
	val := 3
	length := removeElement(nums, val)
	fmt.Println("删除后的长度为：", length)
}

func removeElement(nums []int, val int) int {
	left := 0
	for _, v := range nums {
		if v != val {
			nums[left] = v
			left++
		}
	}
	return left
}

// ShowRemoveDuplicatesV1 删除重复元素
func ShowRemoveDuplicatesV1() {
	nums := []int{1, 1, 2, 3}
	length := removeDuplicatesV2(nums)
	fmt.Println("删除后的长度为：", length)
}

func removeDuplicatesV1(nums []int) int {
	nLen := len(nums)
	if nLen == 0 {
		return 0
	}
	slow := 1
	for fast := 1; fast < nLen; fast++ {
		if nums[fast] != nums[fast-1] {
			nums[slow] = nums[fast]
			slow++
		}
	}
	return slow
}

// ShowRemoveDuplicatesV2 删除出现次数超过两次的元素
func ShowRemoveDuplicatesV2() {
	nums := []int{1, 1, 1, 2, 2, 3}
	length := removeDuplicatesV2(nums)
	fmt.Println("删除后的长度为：", length)
}

func removeDuplicatesV2(nums []int) int {
	slow, fast := 2, 2
	if len(nums) <= 2 {
		return len(nums)
	}
	for fast < len(nums) {
		if nums[slow-2] != nums[fast] {
			nums[slow] = nums[fast]
			slow++
		}
		fast++
	}
	return slow
}
