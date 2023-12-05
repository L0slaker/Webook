package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// GenerateJWT 生成JWT Token
func GenerateJWT(user User) (string, error) {
	//将用户信息 加入 JWT 负载中
	claims := jwt.MapClaims{
		"userId":   user.ID,                               //ID
		"username": user.Username,                         //Username
		"email":    user.Email,                            //Email
		"exp":      time.Now().Add(time.Hour * 24).Unix(), //过期时间
		"iat":      time.Now().Unix(),                     //发布时间
		"nbf":      time.Now().Unix(),                     //生效时间
	}

	//生成 JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//定义签名的密钥
	secret := []byte("secret-key")

	//签名 JWT Token
	signedToken, err := token.SignedString(secret)

	return signedToken, err
}

// ValidateJWT 获取、解析和验证 JWT Token
func ValidateJWT(c *gin.Context) (jwt.MapClaims, error) {
	//1.获取请求中的 JWT Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("缺少Authorization头部，未获得授权")
	}

	//2.处理获取到的Bearer Token字符串
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	//3.定义签名的密钥
	secret := []byte("secret-key")

	//4.验证 JWT Token
	var claims jwt.MapClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, errors.New("fail to parse JWT Token")
	}

	//5.检查 JWT 中保存的信息
	if !token.Valid {
		return nil, errors.New("JWT Token 无效")
	}

	//6.检查过期
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}
