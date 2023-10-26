package web

import (
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/logger"
	"github.com/gin-gonic/gin"
)

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/users")
	group.POST("/signup", u.SignUp)
	group.POST("/login",
		ginx.WrapBodyV1[LoginReq](u.LoginJWT))
	group.POST("/edit", u.EditJWT)
	group.GET("/profile", u.ProfileJWT)
	group.POST("/login_sms/send/code", u.SendLoginSMSCode)
	//group.POST("/login_sms", u.LoginSMS)
	group.POST("/login_sms",
		ginx.WrapBody[LoginSMSReq](u.l.With(logger.String("method", "login_sms")), u.LoginSMS))
	group.POST("/refresh_token", u.RefreshToken)
	group.GET("/exit", u.ExitJWT)
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("/edit", a.Edit)
	group.POST("/publish", a.Publish)
	group.POST("/withdraw", a.Withdraw)
	// 创作者的查询接口
	group.POST("/list",
		ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](a.List))
	// 创作者查看文章详情
	group.GET("/detail/:id",
		ginx.WrapBody[ijwt.UserClaims](a.l.With(logger.String("method", "detail")), a.Detail))

	pub := group.Group("/pub")
	pub.GET("/:id", a.PubDetail)
}
