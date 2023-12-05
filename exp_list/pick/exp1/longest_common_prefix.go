package main

import "fmt"

func ShowLongestCommonPrefix() {
	s := []string{"lisa", "like", "liming"}
	//解法一：横向扫描
	res1 := longestCommonPrefixV1(s)
	//解法二：纵向扫描
	res2 := longestCommonPrefixV2(s)
	//解法三：分治
	res3 := longestCommonPrefixV3(s)
	//解法四：二分查找
	fmt.Printf("方法一_横向扫描：%s 的最长公共前缀是 %s \n", s, res1)
	fmt.Printf("方法二_纵向扫描：%s 的最长公共前缀是 %s \n", s, res2)
	fmt.Printf("方法三_分治：%s 的最长公共前缀是 %s \n", s, res3)
}
func longestCommonPrefixV1(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	count := len(strs)
	for i := 0; i < count; i++ {
		prefix = loopV1(prefix, strs[i])
		if len(prefix) == 0 {
			break
		}
	}
	return prefix
}
func loopV1(s1, s2 string) string {
	length := 0
	if len(s1) < len(s2) {
		length = len(s1)
	} else {
		length = len(s2)
	}
	index := 0
	for index < length && s1[index] == s2[index] {
		index++
	}
	return s1[:index]
}
func longestCommonPrefixV2(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	for i := 0; i < len(strs[0]); i++ {
		for j := 0; j < len(strs); j++ {
			if i == len(strs[j]) || strs[j][i] != strs[0][i] {
				return strs[0][:i]
			}
		}
	}
	return strs[0]
}
func longestCommonPrefixV3(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	var loopV3 func(int, int) string
	loopV3 = func(start, end int) string {
		if start == end {
			return strs[start]
		}
		mid := (start + end) / 2
		left, right := loopV3(start, mid), loopV3(mid+1, end)
		minLength := min(len(left), len(right))
		for i := 0; i < minLength; i++ {
			if left[i] != right[i] {
				return left[:i]
			}
		}
		return left[:minLength]
	}
	return loopV3(0, len(strs)-1)
}
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
