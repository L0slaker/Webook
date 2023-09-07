package ioc

import (
	"Prove/webook/internal/web"
	"Prove/webook/internal/web/middleware"
	"Prove/webook/pkg/ginx/middleware/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitEngine(middlewares []gin.HandlerFunc, handler *web.UserHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middlewares...)
	handler.RegisterRoutes(r)
	return r
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHandler(),
		loginHandler(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		// 加上 ExposeHeaders，前端才能拿到
		ExposeHeaders: []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

func loginHandler() gin.HandlerFunc {
	//return middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/signup").
	//	IgnorePaths("/users/login").
	//	IgnorePaths("/users/login_sms/send/code").
	//	IgnorePaths("/users/login_sms").
	//	Build()
	return middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/send/code").
		IgnorePaths("/users/login_sms").
		Build()
}
