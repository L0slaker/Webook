package main

import "fmt"

// ShowStrStr 找出字符串中第一个匹配项的下标
func ShowStrStr() {
	haystack := "sadbutsad"
	needle := "sad"
	result := strStr(haystack, needle)
	fmt.Println("字符串一: ", haystack)
	fmt.Println("匹配项: ", needle)
	fmt.Println("字符串中第一个匹配项的下标: ", result) // 输出: 0

	haystack2 := "leetcode"
	needle2 := "leeto"
	result2 := strStr(haystack2, needle2)
	fmt.Println("字符串二: ", haystack2)
	fmt.Println("匹配项: ", needle2)
	fmt.Println("字符串中第一个匹配项的下标: ", result2) // 输出: -1
}

func strStr(haystack, needle string) int {
	h, n := len(haystack), len(needle)
outer:
	for i := 0; i+n <= h; i++ {
		for j := range needle {
			if haystack[i+j] != needle[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
