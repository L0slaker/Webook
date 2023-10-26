package ratelimit

import (
	"Prove/webook/internal/service/sms"
	"Prove/webook/pkg/ratelimit"
	"context"
	"fmt"
)

// 装饰器模式

const key = "sms:tencent"

var ErrLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc   sms.Service
	limit ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limit ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:   svc,
		limit: limit,
	}
}

func (s *RatelimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 可以在这里加一些限流或其他新特性
	limited, err := s.limit.Limit(ctx, key)
	if err != nil {
		// 系统错误
		// 可以限流：下游不可靠
		// 可以不限流：下游可靠，业务可用性高，尽量容错策略
		return fmt.Errorf("短信服务判断是否限流出现问题：%w", err)
	}
	if limited {
		return ErrLimited
	}
	err = s.svc.Send(ctx, tplId, args, numbers...)
	// 可以在这里加一些新特性
	return err
}
