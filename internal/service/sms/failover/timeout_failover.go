package failover

import (
	"Prove/webook/internal/service/sms"
	"context"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	idx int32
	// 连续超时次数
	cnt int32
	// 连续超时次数阈值
	threshold int32
	svcs      []sms.Service
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt >= t.threshold {
		// 切换 id
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 往后挪一位，将超时次数设为 0
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现并发，别人换成功了
		//idx = newIdx
		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 连续状态被打断
		atomic.StoreInt32(&t.cnt, 0)
	case context.DeadlineExceeded:
		// 超时
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 未知错误，返回错误或换下一个
	}
	return err
}
