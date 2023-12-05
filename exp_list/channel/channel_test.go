package channel

import (
	"strconv"
	"sync"
	"testing"
)

func TestChannel(t *testing.T) {
	//// 这样声明了channel，但是没有初始化；直接操作读写都会崩溃
	//var ch1 chan int
	//// 一般用于作信号
	//var ch2 chan struct{}

	// 这样的既声明了channel，同时也初始化了，可以直接操作
	// 没有容量的channel
	//ch3 := make(chan int)
	// 带容量的channel，但容量是固定的
	ch4 := make(chan int, 2)

	ch4 <- 123456
	ch4 <- 789
	val, ok := <-ch4
	t.Log(val, ok)
	close(ch4)

	val, ok = <-ch4
	t.Log(val, ok)
}

func TestForLoop(t *testing.T) {
	ch := make(chan string)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- "this is " + strconv.Itoa(i)
		}
		close(ch)
	}()
	for val := range ch {
		t.Log(val)
	}
	t.Log("发送完毕！")
}

func TestSelect(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() {
		ch1 <- 123
		ch2 <- 456
	}()
	select {
	case val := <-ch1:
		t.Log("case1-ch1：", val)
		val = <-ch2
		t.Log("case1-ch2：", val)
	case val := <-ch2:
		t.Log("ch2：", val)
		val = <-ch1
		t.Log("ch1：", val)
	}
}

type MyStruct struct {
	ch chan struct{}
	// 保证在多个goroutine并发调用Close方法时，只有第一个调用会真正关闭通道
	closeOnce sync.Once
}

func (m *MyStruct) Close() error {
	m.closeOnce.Do(func() {
		close(m.ch)
	})
	return nil
}
