package ioc

import (
	"Prove/webook/pkg/logger"
	"go.uber.org/zap"
)

// InitLogger 全局共享的日志
func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
