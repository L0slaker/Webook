package aliyun_v1

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"os"
	"testing"
)

func TestService_SendSms(t *testing.T) {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"), os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"))
	client, err := dysmsapi.NewClientWithOptions("cn-hangzhou", config, credential)
	if err != nil {
		panic("启动客户端失败！")
	}

	signName := "阿里云短信测试"
	templateCode := "SMS_154950909"
	phoneNumber := "13509516520"
	templateParam := "1234"

	svc := NewService(client, signName)
	err = svc.Send(context.Background(), templateCode, []string{templateParam}, phoneNumber)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("发送成功！")
}
