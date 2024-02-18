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
	ioc.InitSRC,
	ioc.InitDST,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
)

var interactiveSvcProvider = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorWeb,
	ioc.InitMigradatorProducer,
	ioc.InitFixDataConsumer,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		migratorProvider,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
