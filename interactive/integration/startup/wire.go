//go:build wireinject

package startup

import (
	"Prove/webook/interactive/grpc"
	"Prove/webook/interactive/repository"
	"Prove/webook/interactive/repository/cache"
	"Prove/webook/interactive/repository/dao"
	"Prove/webook/interactive/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	InitRedis,
	InitTestDB,
	InitLog,
)

var interactiveSvcProvider = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(thirdProvider, interactiveSvcProvider, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
