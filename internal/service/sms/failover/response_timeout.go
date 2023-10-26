package failover

import (
	"Prove/webook/internal/service/sms"
	"Prove/webook/internal/service/sms/ratelimit"
	"context"
	"sync/atomic"
	"time"
)

type ResponseTimeoutFailoverSMSService struct {
	idx int32
	// 响应超时次数
	timeoutCnt int32
	// 总响应次数
	totalCnt int32
	// 使用int32来存储响应时间，以进行原子操作
	responseTimes []int32
	svcs          []sms.Service
}

func NewResponseTimeoutFailoverSMSService(svcs []sms.Service) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs: svcs,
	}
}

func (t *ResponseTimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	startTime := time.Now()
	idx := atomic.LoadInt32(&t.idx)
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 连续状态被打断
		atomic.StoreInt32(&t.timeoutCnt, 0)
		atomic.StoreInt32(&t.totalCnt, 0)

		// 记录每个服务商的响应时间
		responseTime := time.Since(startTime)
		atomic.StoreInt32(&t.responseTimes[idx], int32(responseTime))

		if responseTime > time.Second*1 {
			// 记录超过1s的响应
			atomic.AddInt32(&t.timeoutCnt, 1)
		}
		atomic.AddInt32(&t.totalCnt, 1)
		// 超时比率
		timeoutRate := t.timeoutCnt / t.totalCnt

		// 有10次响应以上且超时比率>20% -> 认为服务商崩溃，需要更换服务
		if t.totalCnt > 10 && float64(timeoutRate) > 0.2 {
			go func() {
				changeService(idx, t)
			}()
		}

	case ratelimit.ErrLimited:
		// 触发了限流
		go func() {
			changeService(idx, t)
		}()
	default:
	}
	return err
}

func changeService(idx int32, t *ResponseTimeoutFailoverSMSService) {
	newIdx := (idx + 1) % int32(len(t.svcs))
	if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
		// 往后挪一位，将超时次数设为 0
		atomic.StoreInt32(&t.timeoutCnt, 0)
		atomic.StoreInt32(&t.totalCnt, 0)
	}
	// else 就是出现并发，别人换成功了;idx = newIdx
	idx = atomic.LoadInt32(&t.idx)
}
