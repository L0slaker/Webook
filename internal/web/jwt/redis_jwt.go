package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	AccessTokenKey  = []byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz")
	RefreshTokenKey = []byte("E9F187B2CA18CA67CF6F67BAA23F1LBJ")
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWT(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = r.SetRefreshToken(ctx, uid, ssid)
	return err
}

func (r *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(AccessTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenString)
	return nil
}

func (r *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
		//UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(RefreshTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenString)
	return nil
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet("claims").(*UserClaims)

	// 将废弃 token 放入本地缓存或redis
	return r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid),
		"", time.Hour*24*7).Err()
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		// 说明 session 不存在
		return nil
	case nil:
		if val == 0 {
			return nil
		}
		return errors.New("session 已经无效了！")
	default:
		return err
	}
}

func (r *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
	//UserAgent string
}
