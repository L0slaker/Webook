package web

import (
	"Prove/webook/internal/service"
	"Prove/webook/internal/service/sms/oauth2/wechat"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 1.用于构造跳转到微信的 URL 接口
// 2.用于处理微信跳转回来的请求

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserAndService
	jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserAndService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/oauth2/wechat")
	group.GET("/authurl", o.AuthURL)
	group.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	url, err := o.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "构造扫码URL失败！",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	info, err := o.svc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}

	u, err := o.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}

	// 设置登陆态
	err = o.setJWTToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "微信登陆成功！",
	})
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
