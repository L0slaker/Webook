package time_ticker

import (
	"context"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	// 一秒的定时器
	tm := time.NewTicker(time.Second)
	defer tm.Stop()
	// tm.C 是一个只读的通道
	for now := range tm.C {
		t.Log(now)
	}
}

func TestTimeCTX(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tm := time.NewTicker(time.Second)
	defer tm.Stop()
	for {
		select {
		case now := <-tm.C:
			t.Log(now.String())
		case <-ctx.Done():
			t.Log("超时或被取消了！")
			return
		}
	}
}
