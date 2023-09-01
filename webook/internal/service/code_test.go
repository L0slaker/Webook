package service

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"os"
	"testing"
)

//LTAI5tPd2puB2DMpFKyNupGP
//HHCb1QjkxWjJ2bIeL5tqwsJKMIOxHr

func TestAliyun_Send(t *testing.T) {
	//accessKey := "LTAI5tPd2puB2DMpFKyNupGP"
	//accessSecret := "HHCb1QjkxWjJ2bIeL5tqwsJKMIOxHr"

	config := sdk.NewConfig()

	// Please ensure that the environment variables ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET are set.
	credential := credentials.NewAccessKeyCredential(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"), os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"))
	/* use STS Token
	credential := credentials.NewStsTokenCredential(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"), os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"), os.Getenv("ALIBABA_CLOUD_SECURITY_TOKEN"))
	*/
	client, err := dysmsapi.NewClientWithOptions("cn-hangzhou", config, credential)
	if err != nil {
		panic(err)
	}

	request := dysmsapi.CreateSendSmsRequest()

	request.Scheme = "https"

	request.SignName = "阿里云短信测试"
	request.TemplateCode = "SMS_154950909"
	request.PhoneNumbers = "13509516520"
	request.TemplateParam = "{\"code\":\"123456\"}"

	response, err := client.SendSms(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
}
