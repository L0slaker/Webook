package context

import (
	"context"
	"testing"
	"time"
)

type Key1 struct{}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, Key1{}, "value1")
	val := ctx.Value(Key1{})
	t.Log(val)
}

func TestContextSub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	subCtx, _ := context.WithCancel(ctx)

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	go func() {
		// 监听 subCtx 的结束信号
		t.Log("等待信号...")
		<-subCtx.Done()
		t.Log("收到信号...")
	}()
	time.Sleep(time.Second * 10)
}