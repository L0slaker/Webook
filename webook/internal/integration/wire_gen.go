// Code generated by Wire. DO NOT EDIT.
//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package integration

import (
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	"Prove/webook/ioc"
	"github.com/gin-gonic/gin"
)

import (
	_ "github.com/gin-contrib/sessions/redis"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewUserInfoDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserInfoRepository(userDAO, userCache)
	userAndService := service.NewUserService(userRepository)
	codeCache := ioc.InitCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeAndService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userAndService, codeAndService)
	engine := ioc.InitEngine(v, userHandler)
	return engine
}