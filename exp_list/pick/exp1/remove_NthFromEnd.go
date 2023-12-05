package main

type ListNode struct {
	Val  int
	Next *ListNode
}

func removeNthFromEnd(head *ListNode, n int) *ListNode {
	length := gentLength(head)

	dNode := &ListNode{0, head}

	cur := dNode
	for i := 0; i < length-n; i++ {
		cur = cur.Next
	}
	cur.Next = cur.Next.Next
	return dNode.Next
}

func gentLength(head *ListNode) int {
	length := 0
	for ; head != nil; head = head.Next {
		length++
	}
	return length
}

func rotateRight(head *ListNode, k int) *ListNode {
	if k == 0 || head == nil || head.Next == nil {
		return head
	}
	n := 1
	iter := head
	for iter.Next != nil {
		iter = iter.Next
		n++
	}
	add := n - k%n
	if add == n {
		return head
	}

	iter.Next = head
	for add > 0 {
		iter = iter.Next
		add--
	}
	ret := iter.Next
	iter.Next = nil
	return ret
}
