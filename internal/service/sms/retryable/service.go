package retryable

import (
	"Prove/webook/internal/service/sms"
	"context"
	"errors"
)

// RetryableService 注意并发问题
type RetryableService struct {
	svc sms.Service
	// 重试值
	retryCnt int
}

func NewRetryableService(svc sms.Service, retryCnt int) sms.Service {
	return &RetryableService{
		svc:      svc,
		retryCnt: retryCnt,
	}
}

func (s *RetryableService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tplId, args, numbers...)
	// 考虑只重试 10 次
	for err != nil && s.retryCnt < 10 {
		err = s.svc.Send(ctx, tplId, args, numbers...)
		s.retryCnt++
	}
	return errors.New("重试都失败了！")
}
