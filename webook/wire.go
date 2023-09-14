//go:build wireinject

package main

import (
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	"Prove/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 初始化 最基础的两个第三方依赖
		ioc.InitDB, ioc.InitRedis,
		// 初始化 UserDAO,UserCahce,CodeCache
		dao.NewUserInfoDAO, cache.NewUserCache, ioc.InitCodeCache,
		// 初始化 UserInfoRepository,CachedCodeRepository
		repository.NewUserInfoRepository, repository.NewCodeRepository,
		// 初始化 UserService,CodeService,SmsService
		service.NewUserService, service.NewCodeService, ioc.InitSMSService, ioc.InitWechatService,
		// 初始化 UserHandler
		web.NewUserHandler, web.NewOAuth2WechatHandler,
		// 初始化 Middleware
		ioc.InitMiddlewares,
		// 初始化 Engine
		ioc.InitEngine,
	)
	return new(gin.Engine)
}
