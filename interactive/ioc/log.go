package ioc

import (
	"Prove/webook/pkg/logger"
	"go.uber.org/zap"
)

func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	//cfg := zap.NewDevelopmentConfig()
	//err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	//l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
