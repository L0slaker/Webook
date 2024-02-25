package main

import (
	"fmt"
	"reflect"
)

func main() {
	// 示例 1: Kind 和 Type
	var x int
	typeOfX := reflect.TypeOf(x)
	kindOfX := typeOfX.Kind()

	fmt.Printf("TypeOf(x): %v\n", typeOfX)
	fmt.Printf("Kind of x: %v\n", kindOfX)

	// 示例 2: Elem
	ptr := reflect.ValueOf(&x)
	value := ptr.Elem() // 获取指针指向的值
	fmt.Printf("TypeOf(&x): %v\n", ptr.Type())
	fmt.Printf("Value: %v\n", value)

	// 示例 3: Interface
	var y interface{}
	y = 42
	valueOfY := reflect.ValueOf(y)
	valueAsInt := valueOfY.Interface().(int) // 将空接口转为具体类型

	fmt.Printf("TypeOf(y): %v\n", reflect.TypeOf(y))
	fmt.Printf("Value of y: %v\n", valueOfY)
	fmt.Printf("Value of y as int: %v\n", valueAsInt)
}

//type Person struct {
//	Name  string
//	Age   int
//	Email string
//}
//
//func main() {
//	// 创建一个 Person 类型的实例
//	personInstance := Person{
//		Name:  "John Doe",
//		Age:   30,
//		Email: "john@example.com",
//	}
//
//	// 获取 reflect.Type 对象
//	personType := reflect.TypeOf(personInstance)
//	fmt.Println("Type of personInstance:", personType)
//
//	// 获取 reflect.Value 对象
//	personValue := reflect.ValueOf(personInstance)
//	fmt.Println("Value of personInstance:", personValue)
//
//	fmt.Println("Type of NumField：", personType.NumField())
//	fmt.Println("Values of NumField：", personValue.NumField())
//
//	// 使用 reflect.Type 获取结构体字段信息
//	for i := 0; i < personType.NumField(); i++ {
//		field := personType.Field(i)
//		fmt.Printf("Field %d: Name=%s, Type=%v\n", i+1, field.Name, field.Type)
//	}
//
//	// 使用 reflect.Value 获取结构体字段值
//	for i := 0; i < personValue.NumField(); i++ {
//		fieldValue := personValue.Field(i)
//		fmt.Printf("Value of Field %d: %v\n", i+1, fieldValue)
//	}
//
//	// 使用 reflect.Value 修改结构体字段值（注意：必须是可修改的字段）
//	if personValue.FieldByName("Age").CanSet() {
//		personValue.FieldByName("Age").SetInt(31)
//	}
//
//	// 输出修改后的结果
//	fmt.Println("Modified personInstance:", personInstance)
//}

//func main() {
//	// 冒泡排序
//	sorted.ShowBubbleSort()
//	// 快速排序
//	sorted.ShowQuickSort()
//	// 归并排序
//	sorted.ShowMergeSort()
//	// 堆排序
//	sorted.ShowHeapSort()
//	// 插入排序
//	sorted.ShowInsertionSort()
//	// 希尔排序
//	sorted.ShowShellSort()
//	// 选择排序
//	sorted.ShowSelectionSort()
//}

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
