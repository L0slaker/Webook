//go:build wireinject

package startup

import (
	artEvent "Prove/webook/internal/events/article"
	"Prove/webook/internal/repository"
	artRepo "Prove/webook/internal/repository/article"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	artDAO "Prove/webook/internal/repository/dao/article"
	"Prove/webook/internal/service"
	interSvc "Prove/webook/internal/service"
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitRedis,
	NewSyncProducer,
	InitKafka,
	InitTestDB, InitLog)

var userSvcProvider = wire.NewSet(
	dao.NewUserInfoDAO,
	cache.NewUserCache,
	repository.NewUserInfoRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	artDAO.NewGORMArticleDAO,
	artRepo.NewArticleRepository,
	service.NewArticleService)

var interactiveSvcProvider = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	interSvc.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,

		InitWechatHandlerConfig,
		artEvent.NewKafkaProducer,
		cache.NewRedisCodeCache,
		repository.NewCodeRepository,
		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitSMSService,

		// 指定啥也不干的 wechat service
		InitPhantomWechatService,
		service.NewCodeService,
		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWT,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitEngine,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler(dao artDAO.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		artEvent.NewKafkaProducer,
		artRepo.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return new(web.ArticleHandler)
}

func InitUserSvc() service.UserAndService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWT)
	return ijwt.NewRedisJWT(nil)
}

func InitInteractiveService() interSvc.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return interSvc.NewInteractiveService(nil, nil)
}
