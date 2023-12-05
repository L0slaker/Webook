package main

import "fmt"

func ShowMaxProfit() {
	prices := []int{7, 1, 5, 3, 6, 4}
	profit := maxProfit(prices)
	fmt.Println("最大利润为", profit)
}

func maxProfit(prices []int) int {
	if len(prices) == 0 {
		return 0
	}

	maxprofit := 0
	minPrice := prices[0]
	for i := 1; i < len(prices); i++ {
		// 如果当前价格比最小价格低，更新最小价格
		if prices[i] < minPrice {
			minPrice = prices[i]
		} else {
			// 如果当前价格减去最小价格比最大利润大，更新最大利润
			if prices[i]-minPrice > maxprofit {
				maxprofit = prices[i] - minPrice
			}
		}
	}
	return maxprofit
}
