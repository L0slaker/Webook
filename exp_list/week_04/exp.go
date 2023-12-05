package week_04

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"math/rand"
	"time"
)

const (
	accessKey    = "MY_ACCESS_KEY"
	secretKey    = "MY_SECRET_KEY"
	signName     = "MY_SIGN_NAME"
	templateCode = "MY_TEMPLATE_CODE"
)

type DatabaseCode struct {
	serverCode string
}

// generateVerificationCode 生成6位的随机验证码
func (dbc *DatabaseCode) generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	min := 000001
	max := 999999
	randomNumber := rand.Intn(max-min+1) + min
	//%06d表示格式化为6位数的十进制数字，并以前导零填充
	formattedNumber := fmt.Sprintf("%06d", randomNumber)

	dbc.serverCode = string(randomNumber)

	return formattedNumber
}

// sendVerificationCode 发送验证码
func sendVerificationCode(phoneNumber string, code string) error {
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", accessKey, secretKey)
	if err != nil {
		return err
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phoneNumber
	request.SignName = signName
	request.TemplateCode = templateCode
	request.TemplateParam = `{"code": "` + code + `"}`

	_, err = client.SendSms(request)
	if err != nil {
		return err
	}

	return nil
}

// verifyVerificationCode 验证验证码
func verifyVerificationCode(userCode string, serverCode string) bool {
	return userCode == serverCode
}

func (dbc *DatabaseCode) handleLogin(phoneNumber string, verificationCode string) {
	serverCode := dbc.serverCode

	if verifyVerificationCode(verificationCode, serverCode) {
		// 验证码匹配，执行登录逻辑
		// TODO: 执行登录逻辑，如创建用户会话、生成访问令牌等
		fmt.Println("登陆成功")
	} else {
		// 验证码不匹配，登陆失败
		fmt.Println("验证码错误")
	}

}
