package web

import (
	"Prove/webook/internal/errs"
	"Prove/webook/internal/service"
	"Prove/webook/internal/service/oauth2/wechat"
	ijwt "Prove/webook/internal/web/jwt"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
)

// 1.用于构造跳转到微信的 URL 接口
// 2.用于处理微信跳转回来的请求

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserAndService
	ijwt.Handler
	stateKey []byte
	cfg      WechatHandlerConfig
}

type WechatHandlerConfig struct {
	Secure bool
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserAndService, cfg WechatHandlerConfig, jwtHandler ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("PGuRsjsp3we94eCAjEnPgQ64uoLY5pzZ"),
		cfg:      cfg,
		Handler:  jwtHandler,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/oauth2/wechat")
	group.GET("/authurl", o.AuthURL)
	group.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()

	url, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatLoginFailed,
			Msg:  "构造扫码URL失败！",
		})
		return
	}

	// 设置 state
	if err = o.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误！",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")

	// 获取并校验 state
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatLoginFailed,
			Msg:  "登陆失败！",
		})
		return
	}

	info, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误！",
		})
		return
	}

	u, err := o.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误！",
		})
		return
	}

	// 设置登陆态
	if err = o.SetLoginToken(ctx, u.Id); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: errs.WechatInternalServerError,
			Msg:  "系统错误！",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "微信登陆成功！",
	})
}

func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenString, err := token.SignedString(o.stateKey)
	if err != nil {
		return err
	}
	// 过期时间十分钟；限制只能在此URL下使用该state；这里HTTPS协议可以选择是否开启
	ctx.SetCookie("jwt-state", tokenString, 600,
		"/oauth2/wechat/callback", "", o.cfg.Secure, true)
	return nil
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")

	tokenString, err := ctx.Cookie("jwt-state")
	if err != nil {
		// 做好监控，记录日志
		return fmt.Errorf("拿不到 state 的cookie，%w", err)
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(tokenString, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.stateKey, nil
	})
	if err != nil || !token.Valid {
		// 做好监控，记录日志
		return fmt.Errorf("token 已经过期，%w", err)
	}
	if sc.State != state {
		// 做好监控，记录日志
		return fmt.Errorf("state 不相等！")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

//
//// OAuth2Handler 统一处理所有的 OAuth2 的
//type OAuth2Handler struct {
//	wechatService
//	dingdingService
//	feishuService
//}
//
//func (o *OAuth2Handler) RegisterRoutes(server *gin.Engine) {
//	group := server.Group("/oauth2")
//	group.GET("/:platform/authurl", o.AuthURL)
//	group.Any("/:platform/callback", o.Callback)
//
//}
//
//func (o *OAuth2Handler) AuthURL(ctx *gin.Context) {
//	platform := ctx.Param("platform")
//	switch platform {
//	case "wechat":
//		o.wechatService.AuthURL
//	case "dingding":
//		o.dingdingService.AuthURL
//	case "feishu":
//		o.feishuService.AuthURL
//	}
//}
//
//func (o *OAuth2Handler) Callback(ctx *gin.Context) {
//
//}
