package main

import (
	"Prove/webook/config"
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/internal/service"
	"Prove/webook/internal/web"
	"Prove/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	r := initWebServer()
	initUser(r, db)

	err := r.Run(":8080")
	if err != nil {
		panic("端口启动失败")
	}
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

func initWebServer() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		//ExposeHeaders: []string{"X-Jwt-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	usingSession(r)

	//usingJWT(r)

	return r
}

func usingSession(r *gin.Engine) {
	store := memstore.NewStore([]byte(""), []byte(""))

	r.Use(sessions.Sessions("ssid", store))
	// 校验
	login := &middleware.LoginMiddlewareBuilder{}
	r.Use(login.CheckLogin())
}

//func usingJWT(r *gin.Engine) {
//
//}

func initUser(r *gin.Engine, db *gorm.DB) {
	da := dao.NewUserInfoDAO(db)
	repo := repository.NewUserInfoRepository(da)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	u.RegisterRoutes(r)
}
