package web

import "github.com/gin-gonic/gin"

func (u *UserHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/users")
	group.POST("/signup", u.SignUp)
	//group.POST("/login", u.Login)
	group.POST("/login", u.LoginJWT)
	//group.POST("/edit", u.Edit)
	group.POST("/edit", u.EditJWT)
	//group.GET("/profile", u.Profile)
	group.GET("/profile", u.ProfileJWT)
	group.GET("/exit", u.Exit)
}

