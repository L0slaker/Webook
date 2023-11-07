//go:build wireinject

package main

import (
	"Prove/webook/internal/events/article"
	"Prove/webook/internal/repository"
	artRepo "Prove/webook/internal/repository/article"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	artDAO "Prove/webook/internal/repository/dao/article"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/ioc"
	"github.com/google/wire"
)

func InitWebServer() *App {
	wire.Build(
		// 初始化第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		ioc.InitKafka, ioc.NewConsumers, ioc.NewSyncProducer,

		// producer & consumer
		//article.NewInteractiveReadEventConsumer,
		// 批量处理
		article.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		// 初始化 dao
		dao.NewUserInfoDAO, artDAO.NewGORMArticleDAO, dao.NewGORMInteractiveDAO,
		cache.NewRedisInteractiveCache, cache.NewUserCache, cache.NewRedisCodeCache,

		// 初始化 repo
		repository.NewUserInfoRepository, repository.NewCodeRepository,
		repository.NewCachedInteractiveRepository, artRepo.NewArticleRepository,

		// 初始化 svc
		service.NewUserService, service.NewCodeService, service.NewArticleService,
		// 基于内存实现
		ioc.InitSMSService, ioc.InitWechatService,

		// 初始化 handler
		web.NewUserHandler, web.NewArticleHandler,
		web.NewOAuth2WechatHandler, ijwt.NewRedisJWT,
		ioc.InitWechatHandlerConfig,

		// 初始化 Middleware
		ioc.InitMiddlewares,
		// 初始化 Engine
		ioc.InitEngine,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
