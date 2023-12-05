package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

func Register(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		//1.解析用户输入的数据
		var u User
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}

		//2.判断输入的信息是否为空
		if u.Username == "" || u.Password == "" || u.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "传入的用户名、密码和邮箱都不能为空！",
			})
			return
		}

		//3.按照要求将密码加密
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to generate password",
			})
			return
		}

		//4.存入数据库
		u.Password = string(hashedPassword)
		if res := db.Create(&u); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": res.Error.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"msg": fmt.Sprintf("%v,注册成功", u.Username),
		})
	}
}

func Login(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var u User
		//1.解析数据
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		//2.判断用户名密码是否为空
		if u.Username == "" || u.Password == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "传入的用户名、密码不能为空",
			})
			return
		}

		//3.查询数据库是否有相关记录
		var user User
		if res := db.Where("username = ?", u.Username).First(&user); res.Error != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "传入的用户名或密码有误",
			})
			return
		}

		//4.若有相关用户，则校验密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "传入的用户名或密码有误!",
			})
			return
		}

		//5.生成 JWT Token
		signedToken, err := GenerateJWT(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sign the JWT",
			})
		}

		//重定向到登陆后的页眉
		//c.Redirect(http.StatusFound, "/article")

		//返回token给用户授权
		c.JSON(http.StatusOK, gin.H{
			"token": signedToken,
		})
	}
}

func Logout() func(c *gin.Context) {
	return func(c *gin.Context) {
		//1.获取、解析和验证 JWT Token
		claims, err := ValidateJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err,
			})
		}

		//2.将 JWT Token 加入到 Token 黑名单列表中
		blackList := make(map[string]bool)
		//jti: JWT Token ID
		tokenID := claims["jti"].(string)
		blackList[tokenID] = true

		//c.Redirect(http.StatusOK, "/login")
		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
	}
}

func AuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		//1.获取、解析和验证 JWT Token
		claims, err := ValidateJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err,
			})
		}

		//7.验证通过，将用户信息存入 Context 中
		c.Set("user_id", uint(claims["user_id"].(float64)))

		//继续处理请求
		c.Next()
	}
}
