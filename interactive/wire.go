//go:build wireinject

package main

import (
	"Prove/webook/interactive/events"
	"Prove/webook/interactive/grpc"
	"Prove/webook/interactive/ioc"
	"Prove/webook/interactive/repository"
	"Prove/webook/interactive/repository/cache"
	"Prove/webook/interactive/repository/dao"
	"Prove/webook/interactive/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitKafka,
)

var interactiveSvcProvider = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		ioc.NewConsumers,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
