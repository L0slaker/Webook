package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	db := ConnectDB()
	config := cors.DefaultConfig()

	//来源跨域
	config.AllowOrigins = []string{"http://localhost:7777"}
	//是否允许发送 Cookie 和 HTTP 认证信息
	config.AllowCredentials = true
	//允许的 HTTP 方法
	config.AllowMethods = []string{"POST", "GET", "PUT", "OPTIONS", "DELETE"}

	r.Use(cors.New(config))

	//1.注册
	r.POST("/register", Register(db))
	//2.登录
	r.POST("/login", Login(db))

	//3.对登录之后可进行的操作进行分组,并应用 middleware 解决登录态问题
	articleAPI := r.Group("/article")
	{
		//3.1创建文章
		articleAPI.POST("/create", AuthMiddleware(db), CreateArticle(db))
		//3.2删除文章
		articleAPI.DELETE("/delete/:id", AuthMiddleware(db), DeleteArticle(db))
		//3.3更新文章
		articleAPI.PUT("/update/:id", AuthMiddleware(db), UpdateArticle(db))
		//3.4查找文章
		articleAPI.GET("/get/:id", AuthMiddleware(db), GetArticle(db))
		//3.5展示所有文章
		articleAPI.GET("/all", AuthMiddleware(db), GetAllArticles(db))

		//4.退出登录
		articleAPI.GET("/exit", AuthMiddleware(db), Logout(db))
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
