package logger

// Logger 兼容性最好
type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func LoggerExp() {
	var l Logger
	phone := "152xxxx1234"
	l.Info("用户未注册,手机号码是 %s", phone)
}

// LoggerV1 认同参数要有名字
type LoggerV1 interface {
	Info(msg string, args ...Field)
	Debug(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key   string
	Value any
}

func LoggerV1Exp() {
	var l LoggerV1
	phone := "152xxxx1234"
	l.Info("用户未注册", Field{
		Key:   "phone",
		Value: phone,
	})
}

// LoggerV2 有完善代码评审流程的可以使用，不然不建议
type LoggerV2 interface {
	// args 必须是偶数，并且按照 Key-Value, Key-Value 来组织
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func LoggerV2Example() {
	var l LoggerV2
	phone := "152xxxx1234"
	l.Info("用户未注册", "phone", phone)
}
