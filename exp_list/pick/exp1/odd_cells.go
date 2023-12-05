package main

import "fmt"

func ShowOddCells() {
	res := oddCells(2, 3, [][]int{{0, 1}, {1, 1}})
	fmt.Println("奇数值单元格的数目:", res)
}

func oddCells(m int, n int, indices [][]int) int {
	res := make([][]int, m)
	for i := range res {
		res[i] = make([]int, n)
	}

	// 处理增量操作
	for _, v := range indices {
		// 处理列
		for j := range res[v[0]] {
			res[v[0]][j]++
		}
		// 处理行
		for _, row := range res {
			row[v[1]]++
		}
	}

	ans := 0
	// 取出奇数单元格的数目
	for _, row := range res {
		for _, v := range row {
			ans += v % 2
		}
	}
	return ans
}
