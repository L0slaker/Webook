package failover

import (
	"Prove/webook/internal/service/sms"
	"context"
	"errors"
	"log"
	"sync/atomic"
)

type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 轮询，全部都轮询完了还没成功，说明所有服务商都挂了
	// 缺点：每次都从头开始轮询，绝大多数请求会在svcs[0]就成功，负载不均衡；如果svcs有几十个，轮询都很慢
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			// 发送成功
			return nil
		}
		// 输出日志 , 做好监控
		log.Println(err)
	}
	return errors.New("发送失败，所有的服务商都尝试过了")
}

// SendV1 对比 Send 有两个改进点：
// 1.起始 svc 是动态计算的
// 2.区别了错误：context.DeadlineExceeded 和 context.Canceled 可以直接返回
func (f *FailoverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 把下标先后推一位
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			// 调用者超时时间已到
			// 调用者主动取消
			return err
		default:
			// 其他情况，需要打日志
			log.Println(err)
		}
	}
	return errors.New("发送失败，所有的服务商都尝试过了")
}
