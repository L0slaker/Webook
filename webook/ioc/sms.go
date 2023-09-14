package ioc

import (
	"Prove/webook/internal/service/sms"
	"Prove/webook/internal/service/sms/aliyun_v1"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"os"
)

func InitSMSService() sms.Service {
	// 基于内存的实现
	//return memory.NewService()

	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"), os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"))
	client, err := dysmsapi.NewClientWithOptions("cn-hangzhou", config, credential)
	if err != nil {
		panic("启动客户端失败！")
	}
	return aliyun_v1.NewService(client, "阿里云短信测试")
}
