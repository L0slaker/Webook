//go:build wireinject

package startup

import (
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/article"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	article_dao "Prove/webook/internal/repository/dao/article"
	"Prove/webook/internal/service"
	"Prove/webook/internal/service/sms"
	"Prove/webook/internal/service/sms/async"
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewUserInfoDAO,
	cache.NewUserCache,
	repository.NewUserInfoRepository,
	service.NewUserService)
var articleSvcProvider = wire.NewSet(
	article_dao.NewGORMArticleDAO,
	article.NewArticleRepository,
	service.NewArticleService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 初始化第三方依赖
		thirdProvider, userSvcProvider, articleSvcProvider,
		//articlSvcProvider,
		cache.NewRedisCodeCache,
		repository.NewCodeRepository,
		// service 部分
		ioc.InitSMSService, InitPhantomWechatService,
		service.NewCodeService,
		// handler 部分
		web.NewUserHandler, web.NewOAuth2WechatHandler, web.NewArticleHandler,
		InitWechatHandlerConfig, ijwt.NewRedisJWT,
		// gin 的中间件
		ioc.InitMiddlewares,
		// Web 服务器
		ioc.InitEngine,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler(dao article_dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		service.NewArticleService,
		web.NewArticleHandler,
		article.NewArticleRepository,
	)
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

func InitAsyncSmsService(svc sms.Service) *async.SMSService {
	wire.Build(thirdProvider,
		repository.NewAsyncSMSRepository,
		dao.NewGORMAsyncSmsDAO,
		async.NewSMSService,
	)
	return new(async.SMSService)
}
