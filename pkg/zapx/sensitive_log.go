package zapx

import "go.uber.org/zap/zapcore"

// MyCore 装饰器模式——对敏感数据脱敏
type MyCore struct {
	zapcore.Core
}

func (c MyCore) Write(entry zapcore.Entry, fds []zapcore.Field) error {
	for _, fd := range fds {
		if fd.Key == "phone" {
			phone := fd.String
			fd.String = phone[:3] + "****" + phone[7:]
		}
	}
	// 日志条目（zapcore.Entry）和字段（[]zapcore.Field）写入到日志的目标输出位置
	return c.Core.Write(entry, fds)
}
