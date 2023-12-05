package main

import "fmt"

func ShowRotate() {
	nums := []int{1, 2, 3, 4, 5, 6, 7}
	k := 3
	rotateV1(nums, k)
	rotateV2(nums, k)
}

// rotateV1 新数组
func rotateV1(nums []int, k int) {
	fmt.Println("反转前的数组", nums)
	newNums := make([]int, len(nums))
	for i, v := range nums {
		newNums[(i+k)%len(nums)] = v
	}
	copy(nums, newNums)
	fmt.Println("反转后的数组", nums)
}

// rotateV2 数组翻转
func rotateV2(nums []int, k int) {
	fmt.Println("反转前的数组", nums)
	k %= len(nums)
	reverse(nums)
	reverse(nums[:k])
	reverse(nums[k:])
	fmt.Println("反转后的数组", nums)
}

func reverse(a []int) {
	for i, j := 0, len(a); i < j/2; i++ {
		a[i], a[j-i-1] = a[j-i-1], a[i]
	}
}
