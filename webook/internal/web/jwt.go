package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtHandler struct {
}

func (h jwtHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    uid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenString)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	UserAgent string
}
