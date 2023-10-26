package startup

import "Prove/webook/pkg/logger"

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
