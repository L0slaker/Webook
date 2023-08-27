package tencent

import (
	mysms "Prove/webook/internal/service/sms"
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client *sms.Client

	appId    *string
	signName *string
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    ekit.ToPtr[string](appId),
		signName: ekit.ToPtr[string](signName),
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = s.toStringPtr(numbers)
	req.TemplateParamSet = s.toStringPtr(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		// 短信请求验证码为空 或 验证码不对
		if status.Code == nil || (*status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, 原因: %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) SendV1(ctx context.Context, tplId string, args []mysms.NameArg, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = s.toStringPtr(numbers)
	req.TemplateParamSet = slice.Map[mysms.NameArg, *string](args, func(idx int, src mysms.NameArg) *string {
		return &src.Val
	})
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		// 短信请求验证码为空 或 验证码不对
		if status.Code == nil || (*status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, 原因: %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtr(src []string) []*string {
	return slice.Map[string, *string](src, func(id int, src string) *string {
		return &src
	})
}
