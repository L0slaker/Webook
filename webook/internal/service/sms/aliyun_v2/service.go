package aliyun_v2

import (
	"context"
	"encoding/json"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"os"
	"strconv"
	"strings"
)

type Service struct {
	client   *dysms.Client
	signName string
}

func NewService(signName string) *Service {
	client := CreateClient()
	return &Service{
		client:   client,
		signName: signName,
	}
}

func CreateClient() *dysms.Client {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: tea.String(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")),
		// 必填，您的 AccessKey Secret
		AccessKeySecret: tea.String(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	res := &dysms.Client{}
	res, err := dysms.NewClient(config)
	if err != nil {
		panic(err)
	}
	return res
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	argsMap := make(map[string]string, len(args))
	for k, arg := range args {
		argsMap[strconv.Itoa(k)] = arg
	}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	sendSmsRequest := &dysms.SendSmsRequest{
		//SignName:      tea.String("阿里云短信测试"),
		//TemplateCode:  tea.String("SMS_154950909"),
		//PhoneNumbers:  tea.String("13509516520"),
		//TemplateParam: tea.String("{\"code\":\"123456\"}"),
		SignName:      tea.String(s.signName),
		TemplateCode:  tea.String(tplId),
		PhoneNumbers:  tea.String(strings.Join(numbers, ",")),
		TemplateParam: tea.String(string(bCode)),
	}
	var resp *dysms.SendSmsResponse
	resp, err = s.client.SendSms(sendSmsRequest)
	if err != nil {
		return err
	}
	fmt.Println(resp.StatusCode)
	return nil
}
