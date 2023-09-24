package web

import "github.com/gin-gonic/gin"

func (u *UserHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/users")
	group.POST("/signup", u.SignUp)
	//group.POST("/login", u.Login)
	//group.POST("/edit", u.Edit)
	//group.GET("/profile", u.Profile)
	//group.GET("/exit", u.Exit)
	group.POST("/login", u.LoginJWT)
	group.POST("/edit", u.EditJWT)
	group.GET("/profile", u.ProfileJWT)
	group.POST("/login_sms/send/code", u.SendLoginSMSCode)
	group.POST("/login_sms", u.LoginSMS)
	group.POST("/refresh_token", u.RefreshToken)
	group.GET("/exit", u.ExitJWT)
}
