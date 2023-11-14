package ioc

import (
	"Prove/webook/internal/service/oauth2/wechat"
	"Prove/webook/internal/web"
	"Prove/webook/pkg/logger"
	"os"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	// app_id 和 app_secret 都设置在环境变量中
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	svc := wechat.NewService(appId, appSecret, l)
	// 接入 Prometheus 监控
	return wechat.NewPrometheusDecorator(svc, "geekbang_l0slakers",
		"webook", "wechat_resp_time",
		"统计 wechat 服务的性能数据", "my-instance-1")
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
