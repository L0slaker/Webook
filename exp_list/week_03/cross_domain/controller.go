package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
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

func Login(db *gorm.DB, client *redis.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		var u User
		//1.解析数据
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "解析数据失败",
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

		//5.生成Session ID
		sessionID := uuid.New().String()

		userId := strconv.Itoa(int(user.ID))
		//6.将Session ID与用户信息存储到redis数据库中:使用字符串类型存储session id
		err := client.HMSet(c, sessionID, map[string]interface{}{
			"userId":     userId,
			"expiration": time.Now().Add(60 * time.Minute),
		}).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":     err.Error(),
				"error-msg": "存储session id失败",
			})
			return
		}

		//7.将Session ID存储到Cookie中
		c.SetCookie("session_id", sessionID, 60&60*24, "/", "", false, true)

		//8.重定向到登陆后的页眉
		c.Redirect(http.StatusFound, "/article")
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("欢迎回来,%v!", u.Username),
		})
	}
}

func AuthMiddleware(client *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		//1.获取请求中的 Cookie，并从 Cookie 中获取 Session ID
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			//c.Redirect(http.StatusNotFound, "/login")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "获取session ID失败"})
			return
		}

		//2.使用 Session ID 查询数据库中的登录状态信息
		res, err := client.HMGet(c, sessionID, "userId", "expiration").Result()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "获取用户信息失败"})
			return
		}

		userId := res[0].(int)
		expiration := res[1].(int64)

		currentTime := time.Now().Unix()

		//3.判断登录信息是否有效
		if userId == 0 || currentTime > expiration {
			// 会话已过期，进行相应处理
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "会话已过期",
			})
			return
		}

		////4.如果登录信息有效，则将登录信息存入 Context 中
		c.Set("session_id", userId)

		//5.继续处理请求
		c.Next()
	}
}

func Logout(client *redis.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		//通过Context获取session ID
		sessionID := c.Value("session_id")

		//删除用户信息
		if err := client.Del(c, sessionID.(string)).Err(); err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		//c.Redirect(http.StatusOK, "/login")
		c.JSON(http.StatusOK, gin.H{"message": "Logout successfully"})
	}
}
