package logger

type NopLogger struct {
}

func (n *NopLogger) Info(msg string, args ...Field) {
	//TODO implement me
	panic("implement me")
}

func (n *NopLogger) Debug(msg string, args ...Field) {
	//TODO implement me
	panic("implement me")
}

func (n *NopLogger) Warn(msg string, args ...Field) {
	//TODO implement me
	panic("implement me")
}

func (n *NopLogger) Error(msg string, args ...Field) {
	//TODO implement me
	panic("implement me")
}