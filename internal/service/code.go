package service

import (
	"Prove/webook/internal/repository"
	"Prove/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

const codeTplId = "SMS_154950909"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeAndService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

// CodeService 验证码服务
type CodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
	//tplId string
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeAndService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 发送验证码
// biz 区别业务场景
func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// 1.生成验证码
	code := svc.generateCode()
	// 2.放入Redis；选择redis存储的原因，主要是redis是单线程的，不会发生并发问题
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 3.发送验证码
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		//	// 走到这里意味着 Redis 有这个验证码，但是超时了。
		//	// 可以考虑在这里重试，如果需要重试，在初始化的时候需要传入一个重试的实现smsSvc
		return fmt.Errorf("发送短信出现异常 %w", err)
	}
	return err
}

// Verify 验证，保证验证码不会被暴力破解
// 第一个返回值代表验证的正确性，第二个代表系统出的错误
func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	// 六位数，num 在 0~999999 之间
	num := rand.Intn(1000000)
	// 不够六位的，加上前导0补齐
	return fmt.Sprintf("%06d", num)
}
