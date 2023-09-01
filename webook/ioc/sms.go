package ioc

import (
	"Prove/webook/internal/service/sms"
	"Prove/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 暂时是基于内存的实现
	return memory.NewService()
}
