package main

import "Prove/study_list/gramming/sorted"

func main() {
	// 冒泡排序
	sorted.ShowBubbleSort()
	// 快速排序
	sorted.ShowQuickSort()
	// 归并排序
	sorted.ShowMergeSort()
	// 堆排序
	sorted.ShowHeapSort()
	// 插入排序
	sorted.ShowInsertionSort()
	// 希尔排序
	sorted.ShowShellSort()
	// 选择排序
	sorted.ShowSelectionSort()
}

//// 总耗时： 3.5141ms
//func findKthLargest(nums []int, k int) int {
//	n := len(nums)
//	return quickselect(nums, 0, n-1, n-k)
//}
//
//func quickselect(nums []int, l, r, k int) int {
//	if l == r {
//		return nums[k]
//	}
//	partition := nums[l]
//	i := l - 1
//	j := r + 1
//	for i < j {
//		for i++; nums[i] < partition; i++ {
//		}
//		for j--; nums[j] > partition; j-- {
//		}
//		if i < j {
//			nums[i], nums[j] = nums[j], nums[i]
//		}
//	}
//	if k <= j {
//		return quickselect(nums, l, j, k)
//	} else {
//		return quickselect(nums, j+1, r, k)
//	}
//}

// 总耗时： 23.9675622s
//func findKthLargest(nums []int, k int) int {
//	n := len(nums)
//	quickSort(nums, 0, n-1)
//	return nums[n-k]
//}
//
//func quickSort(nums []int, low, high int) {
//	if low < high {
//		pivotIndex := partition(nums, low, high)
//		quickSort(nums, low, pivotIndex-1)
//		quickSort(nums, pivotIndex+1, high)
//	}
//}
//
//func partition(nums []int, low, high int) int {
//	pivot := nums[high]
//	i := low - 1
//	for j := low; j < high; j++ {
//		if nums[j] <= pivot {
//			i++
//			nums[i], nums[j] = nums[j], nums[1]
//		}
//	}
//	nums[i+1], nums[high] = nums[high], nums[i+1]
//	return i + 1
//}
