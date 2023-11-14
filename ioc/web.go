package ioc

import (
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/internal/web/middleware"
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/ginx/middleware/metric"
	logger2 "Prove/webook/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
)

func InitEngine(middlewares []gin.HandlerFunc, handler *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middlewares...)
	handler.RegisterRoutes(r)
	oauth2WechatHdl.RegisterRoutes(r)
	articleHdl.RegisterRoutes(r)
	// 测试专用
	(&web.ObservabilityHandler{}).RegisterRoutes(r)
	return r
}

func InitMiddlewares(redisClient redis.Cmdable, l logger2.LoggerV1, jwtHandler ijwt.Handler) []gin.HandlerFunc {
	//bd := logger.NewMiddlewareBuilder(func(ctx context.Context, al *logger.AccessLog) {
	//	l.Debug("HTTP 请求", logger2.Field{Key: "req", Value: al})
	//}).AllowReqBody(true).AllowRespBody()
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.log_req")
	//	bd.AllowReqBody(ok)
	//})
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_l0slakers",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHandler(),
		(&metric.MiddlewareBuilder{
			Namespace:  "geekbang_l0slakers",
			Subsystem:  "webook",
			Name:       "gin_http",
			Help:       "统计 gin 的HTTP 接口",
			InstanceId: "my-instance-1",
		}).Build(),
		otelgin.Middleware("webook"),
		loginHandler(jwtHandler),
		//bd.Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		// 加上 ExposeHeaders，前端才能拿到
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

func loginHandler(jwtHandler ijwt.Handler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).
		IgnorePaths("/users/signup").
		IgnorePaths("/users/refresh_token").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/send/code").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/oauth2/wechat/authurl").
		IgnorePaths("/oauth2/wechat/callback").
		Build()
}
