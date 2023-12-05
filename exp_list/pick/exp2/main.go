package main

import (
	"fmt"
)

//
//func main() {
//	defer func() {
//		fmt.Print("demo")
//	}()
//	go foo()
//	time.Sleep(time.Second)
//}
//
//func foo() {
//	defer fmt.Print("A")
//	defer fmt.Print("B")
//	fmt.Print("C")
//	panic("demo")
//	defer fmt.Print("D")
//}

func main() {
	foo := func() int {
		defer func() {
			recover()
		}()
		panic("demo")
		return 10
	}
	ret := foo()
	fmt.Println(ret)
	fmt.Println("B")
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func flatten(root *TreeNode) {
	list := preorderTraversal(root)
	fmt.Println(list)
	for i := 1; i < len(list); i++ {
		prev, cur := list[i-1], list[i]
		prev.Left, prev.Right = nil, cur
	}
}

func preorderTraversal(root *TreeNode) []*TreeNode {
	list := []*TreeNode{}
	if root != nil {
		list = append(list, root)
		list = append(list, preorderTraversal(root.Left)...)
		list = append(list, preorderTraversal(root.Right)...)
	}
	return list
}

//func groupAnagrams(strs []string) [][]string {
//	mp := map[[26]int][]string{}
//	for _, str := range strs {
//		cnt := [26]int{}
//		for _, b := range str {
//			cnt[b-'a']++
//			fmt.Println("cnt[b-'a']->", cnt[b-'a'])
//		}
//		mp[cnt] = append(mp[cnt], str)
//		fmt.Println("mp[cnt]->", mp[cnt])
//	}
//	ans := make([][]string, 0, len(mp))
//	for _, v := range mp {
//		ans = append(ans, v)
//	}
//	return ans
//}

//func uniquePathsWithObstacles(obstacleGrid [][]int) int {
//	rows, columns := len(obstacleGrid), len(obstacleGrid[0])
//	dp := make([]int, columns)
//	if obstacleGrid[0][0] == 0 {
//		dp[0] = 1
//	}
//	for i := 0; i < rows; i++ {
//		for j := 0; j < columns; j++ {
//			if obstacleGrid[i][j] == 1 {
//				dp[j] = 0
//				continue
//			}
//			if j-1 >= 0 && obstacleGrid[i][j-1] == 0 {
//				dp[j] += dp[j-1]
//			}
//			println("j->dp[len(dp)-1]:  ", dp[len(dp)-1])
//		}
//		println("i->dp[len(dp)-1]:  ", dp[len(dp)-1])
//	}
//	return dp[len(dp)-1]
//}

//func minimumTotal(triangle [][]int) int {
//	// f[i][j]=min(f[i−1][j−1],f[i−1][j])+c[i][j]
//	h := len(triangle)
//	dp := make([][]int, h)
//	for i := range dp {
//		dp[i] = make([]int, len(triangle[i]))
//	}
//
//	for i := h - 1; i >= 0; i-- {
//		for j := 0; j < len(triangle[i]); j++ {
//			if i == h-1 {
//				dp[i][j] = triangle[i][j]
//			} else {
//				dp[i][j] = min(dp[i+1][j], dp[i+1][j+1]) + triangle[i][j]
//			}
//			println("dp[i][j]: ", dp[i][j])
//		}
//	}
//	return dp[0][0]
//}
//
//func min(a, b int) int {
//	if a > b {
//		return b
//	}
//	return a
//}
