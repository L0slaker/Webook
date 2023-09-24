package jwt

import "github.com/gin-gonic/gin"

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractToken(ctx *gin.Context) string
}
