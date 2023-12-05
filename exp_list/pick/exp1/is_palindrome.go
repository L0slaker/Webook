package main

import "fmt"

func ShowIsPalindrome() {
	x := 53335
	ok := isPalindrome(x)
	fmt.Printf("%d 是回文数吗: ", x)
	fmt.Println(ok)
}
func isPalindrome(x int) bool {
	// 其他情况处理：
	// x<0：这种情况注定会有 - 符号，无法符合回文数的规定
	// x%10==0 && x!=0：这中情况的最后一位是0，则需要第一位数字也为0，只有0能满足此情况
	if x < 0 || (x%10 == 0 && x != 0) {
		return false
	}

	// "/"取商，"%"取余
	revertedNum := 0
	for x > revertedNum {
		revertedNum = revertedNum*10 + x%10
		x /= 10
	}

	if x == revertedNum || x == revertedNum/10 {
		return true
	}
	return false
}
