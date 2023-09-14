package ioc

import (
	"Prove/webook/internal/service/sms/oauth2/wechat"
	"os"
)

func InitWechatService() wechat.Service {
	// app_id 和 app_secret 都设置在环境变量中
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appSecret)
}
