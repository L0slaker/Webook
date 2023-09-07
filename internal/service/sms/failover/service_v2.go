package failover

import (
	"Prove/webook/internal/service/sms"
	"context"
	"errors"
	"log"
)

type FailoverSMSServiceV2 struct {
	svcs []sms.Service
}

func NewFailoverSMSServiceV2(svcs []sms.Service) sms.Service {
	return &FailoverSMSServiceV2{
		svcs: svcs,
	}
}

func (f *FailoverSMSServiceV2) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
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
