package main

import "fmt"

func ShowRomanToInt() {
	// 59
	romeNumber1 := "LIX"
	// 128
	romeNumber2 := "CXXVIII"
	res1 := romanToInt(romeNumber1)
	res2 := romanToInt(romeNumber2)
	fmt.Printf("%s 转为整数：%d  \n", romeNumber1, res1)
	fmt.Printf("%s 转为整数：%d  \n", romeNumber2, res2)
}

var symbolRome = map[byte]int{
	'I': 1,
	'V': 5,
	'X': 10,
	'L': 50,
	'C': 100,
	'D': 500,
	'M': 1000,
}

func romanToInt(s string) int {
	res := 0
	//检索字符串
	for i := range s {
		val := symbolRome[s[i]]
		if i < len(s)-1 && val < symbolRome[s[i+1]] {
			res -= val
		} else {
			res += val
		}
	}
	return res
}
