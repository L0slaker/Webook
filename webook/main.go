package main

import (
	_ "github.com/gin-contrib/sessions/redis"
)

func main() {
	//db := initDB()
	//r := initWebServer()
	//rdb := initRedis()
	//u := initUser(db, rdb)
	//u.RegisterRoutes(r)
	r := InitWebServer()
	err := r.Run(":8081")
	if err != nil {
		panic("端口启动失败")
	}
}

// 依赖注入，迁移
//func initRedis() redis.Cmdable {
//	redisClient := redis.NewClient(&redis.Options{
//		Addr: config.Config.Redis.Addr,
//	})
//	return redisClient
//}

// 依赖注入，迁移
//func initDB() *gorm.DB {
//	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
//	if err != nil {
//		panic(err)
//	}
//	err = dao.InitTables(db)
//	if err != nil {
//		panic(err)
//	}
//	return db
//}

// 依赖注入，迁移
// TypeError: Cannot read properties of undefined (reading 'status')
//func initWebServer() *gin.Engine {
//	r := gin.Default()
//	// 跨域机制
//	r.Use(cors.New(cors.Config{
//		AllowCredentials: true,
//		AllowHeaders:     []string{"Content-Type", "Authorization"},
//		// 加上 ExposeHeaders，前端才能拿到
//		ExposeHeaders: []string{"x-jwt-token"},
//		AllowOriginFunc: func(origin string) bool {
//			if strings.HasPrefix(origin, "http://localhost") {
//				return true
//			}
//			return strings.Contains(origin, "your_company.com")
//		},
//		MaxAge: 12 * time.Hour,
//	}))
//
//	// 限流机制
//	//redisClient := redis.NewClient(&redis.Options{
//	//	Addr: config.Config.Redis.Addr,
//	//})
//	//r.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//
//	//usingSession(r)
//	usingJWT(r)
//	return r
//}

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

// 依赖注入，迁移
//func usingJWT(r *gin.Engine) {
//	//store := memstore.NewStore([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), []byte("EdsbuUneoaqBDWlbLvqP1d1gsDX7GoKH"))
//	//
//	//r.Use(sessions.Sessions("ssid", store))
//	// 校验
//	login := middleware.NewLoginJWTMiddlewareBuilder()
//	r.Use(login.
//		IgnorePaths("/users/signup").
//		IgnorePaths("/users/login").
//		IgnorePaths("/users/login_sms/send/code").
//		IgnorePaths("/users/login_sms").
//		Build())
//}

// 依赖注入，迁移
//func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
//	da := dao.NewUserInfoDAO(db)
//	uc := cache.NewUserCache(rdb)
//	repo := repository.NewUserInfoRepository(da, uc)
//	svc := service.NewUserService(repo)
//
//	//var codeCache *cache.CodeCache
//	//store := &sync.Map{}
//	//localCodeCache := cache.NewLocalCodeCache(store)
//	redisCodeCache := cache.NewRedisCodeCache(rdb)
//	codeRepo := repository.NewCodeRepository(redisCodeCache)
//	smsSvc := memory.NewService()
//	//smsSvc := aliyun_v1.NewService(initClient(), signName)
//	codeSvc := service.NewCodeService(codeRepo, smsSvc)
//
//	u := web.NewUserHandler(svc, codeSvc)
//	return u
//}

//const (
//	accessKeyId  = "LTAI5tPd2puB2DMpFKyNupGP"
//	accessSecret = "HHCb1QjkxWjJ2bIeL5tqwsJKMIOxHr"
//	signName     = "阿里云短信测试"
//)
//
//func initClient() *dysmsapi.Client {
//	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessSecret)
//
//	if err != nil {
//		panic(err)
//	}
//	return client
//}
