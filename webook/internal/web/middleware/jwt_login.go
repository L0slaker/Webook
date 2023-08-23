package middleware

import (
	"Prove/webook/internal/web"
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segments := strings.SplitN(header, " ", 2)
		tokenString := segments[1]
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), nil
		})

		if err != nil || token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 额外校验 UserAgent，增强登陆安全
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题，需要监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新 token / 10s
		if claims.ExpiresAt.Sub(time.Now()) < 50 {
			// 检查过期
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenString, err = token.SignedString([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"))
			if err != nil {
				// 记录日志
				log.Println("续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenString)
		}

		// 为其它功能提供claims
		ctx.Set("claims", claims)
	}
}
