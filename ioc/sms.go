package ioc

import (
	"Prove/webook/internal/service/sms"
	"Prove/webook/internal/service/sms/aliyun_v1"
	"Prove/webook/internal/service/sms/metrics"
	"Prove/webook/internal/service/sms/ratelimit"
	"Prove/webook/internal/service/sms/retryable"
	limiter "Prove/webook/pkg/ratelimit"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 基于内存的实现
	//return memory.NewService()

	// 基于阿里云v1的实现
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(
		os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
		os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"))
	client, err := dysmsapi.NewClientWithOptions("cn-hangzhou", config, credential)
	if err != nil {
		panic("启动客户端失败！")
	}

	// 限流机制
	svc := ratelimit.NewRatelimitSMSService(aliyun_v1.NewService(client, "阿里云短信测试"),
		limiter.NewRedisSlideWindowLimiter(cmd, time.Second, 100))

	// 超时重试机制

	// 日志机制

	// 监控
	svc = metrics.NewPrometheusDecorator(svc)

	// 重试机制
	return retryable.NewRetryableService(svc, 3)
}
