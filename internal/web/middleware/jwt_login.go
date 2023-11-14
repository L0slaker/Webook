package middleware

import (
	ijwt "Prove/webook/internal/web/jwt"
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func NewLoginJWTMiddlewareBuilder(jwtHandler ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHandler,
	}
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		tokenString := l.ExtractToken(ctx)
		claims := &ijwt.UserClaims{}
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

		// 验证ssid
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			// 系统错误或用户已经退出登录
			// 这里同样也可以考虑在 redis 崩溃后就不去校验是否主动退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 为其它功能提供claims
		ctx.Set("claims", claims)
	}
}
