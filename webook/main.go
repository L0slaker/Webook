package main

import (
	"Prove/webook/config"
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	"Prove/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	_ "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	r := initWebServer()
	rdb := initRedis()
	u := initUser(db, rdb)
	u.RegisterRoutes(r)

	err := r.Run(":8081")
	if err != nil {
		panic("端口启动失败")
	}
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	return redisClient
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// TypeError: Cannot read properties of undefined (reading 'status')
func initWebServer() *gin.Engine {
	r := gin.Default()
	// 跨域机制
	r.Use(cors.New(cors.Config{
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
	}))

	// 限流机制
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//r.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	//usingSession(r)
	usingJWT(r)
	return r
}

//func usingSession(r *gin.Engine) {
//	// 基于cookie
//	//store := cookie.NewStore([]byte("secret"))
//	// 基于内存，一般用于单实例部署
//	// 随机生成的32位密码，第一个参数是 authentication key(身份认证)，第二个参数是 encryption key(数据加密)
//	//store := memstore.NewStore([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), []byte("EdsbuUneoaqBDWlbLvqP1d1gsDX7GoKH"))
//	// 基于redis，可用于多实例部署
//	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//		[]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), []byte("EdsbuUneoaqBDWlbLvqP1d1gsDX7GoKH"))
//	if err != nil {
//		panic(err)
//	}
//
//	//myStore := &sqlx_store.Store{}
//
//	r.Use(sessions.Sessions("ssid", store))
//	// 校验
//	login := middleware.NewLoginMiddlewareBuilder()
//	r.Use(login.IgnorePaths("/users/signup").IgnorePaths("/users/login").Build())
//}

func usingJWT(r *gin.Engine) {
	//store := memstore.NewStore([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), []byte("EdsbuUneoaqBDWlbLvqP1d1gsDX7GoKH"))
	//
	//r.Use(sessions.Sessions("ssid", store))
	// 校验
	login := middleware.NewLoginJWTMiddlewareBuilder()
	r.Use(login.IgnorePaths("/users/signup").IgnorePaths("/users/login").Build())
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	da := dao.NewUserInfoDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserInfoRepository(da, uc)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	//u.RegisterRoutes(r)
	return u
}
