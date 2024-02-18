package main

import (
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/grpcx"
	"Prove/webook/pkg/saramax"
)

// App 控制 main 函数的声明周期
type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	webAdmin  *ginx.Server
}
