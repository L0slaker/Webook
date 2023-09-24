package aliyun_v1

import (
	"Prove/webook/internal/service/sms"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"go.uber.org/zap"
	"strings"
)

type Service struct {
	client   *dysmsapi.Client
	signName string
}

func NewService(client *dysmsapi.Client, signName string) *Service {
	return &Service{
		client:   client,
		signName: signName,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "http"
	// 阿里云手机号由字符串逗号间隔
	req.PhoneNumbers = strings.Join(numbers, ",")
	req.SignName = s.signName

	// 需要一个map
	argsMap := make(map[string]string, len(args))
	for _, arg := range args {
		argsMap["code"] = arg
	}

	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	//req.TemplateParam = string(bCode)
	req.TemplateCode = biz

	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	zap.L().Debug("发送短信", zap.Error(err),
		zap.Any("req", req), zap.Any("resp", resp))
	if err != nil {
		return err
	}

	if resp.Code != "OK" {
		return fmt.Errorf("发送失败，code: %s, 原因：%s",
			resp.Code, resp.Message)
	}
	return nil
}

func (s *Service) SendV1(ctx context.Context, tplId string, args []sms.NameArg, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "http"
	// 阿里云手机号由字符串逗号间隔
	req.PhoneNumbers = strings.Join(numbers, ",")
	req.SignName = s.signName

	// 需要一个map
	argsMap := make(map[string]string, len(args))
	for _, arg := range args {
		argsMap[arg.Name] = arg.Val
	}

	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	req.TemplateCode = tplId

	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	if err != nil {
		return err
	}

	if resp.Code != "OK" {
		return fmt.Errorf("发送失败，code: %s, 原因：%s",
			resp.Code, resp.Message)
	}
	return nil
}
