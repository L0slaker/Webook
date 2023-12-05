package sorted

import "fmt"

// heapSort 升序使用大顶堆，降序使用小顶堆
// 1.把无序数组构建成二叉堆
// 2.循环删除堆顶元素，移到集合尾部，调节堆产生新的堆顶
func heapSort(s []int) []int {
	length := len(s)
	// 对于完全二叉树来说，最后一个非叶子节点的索引是 len(s)/2 - 1
	for i := length/2 - 1; i >= 0; i-- {
		heapify(s, length, i)
	}

	// 循环删除堆顶元素，移到集合尾部，调节堆产生新的堆顶
	for i := length - 1; i > 0; i-- {
		s[0], s[i] = s[i], s[0]
		heapify(s, i, 0)
	}
	return s
}

// 维护最大堆的函数
// 比较其左右节点中最大的值，是否比最后一个非叶子节点大，若大则交换位置
// 再找到下一个非叶子节点，继续比较和交换
func heapify(s []int, length, root int) {
	// 初始化根节点为最大值
	largest := root
	// 左右子树的索引
	left := 2*root + 1
	right := 2*root + 2

	// 如果左、右子节点比根节点大，则更新最大值的索引
	if left < length && s[left] > s[largest] {
		largest = left
	}
	if right < length && s[right] > s[largest] {
		largest = right
	}
	// 如果最大值不是根节点，则交换并继续向下维护最大堆
	if largest != root {
		s[root], s[largest] = s[largest], s[root]
		heapify(s, length, largest)
	}
}

func ShowHeapSort() {
	fmt.Println(">堆排序<")
	s := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("排序前", s) // 排序前
	arr := quickSort(s)
	// 排序后
	fmt.Println("排序后", arr) // 排序前
}
