package web

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/service"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/logger"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
)

const biz = "login"

const (
	//邮箱规则
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	//密码规则：至少1个大小写字母、一个数字、一个特殊符号，长度至少为8个字符
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	//昵称规则：长度为2到20个字符，可以是字母（大小写）、数字或下划线。
	nicknameRegexPattern = `^[A-Za-z0-9_]{2,20}$`
	//生日规则：以19或20开头的四位数字年份、'-'分隔的两位数字的月份，范围从01到12、'-'分隔的两位数字的日期，范围从01到31,考虑了不同月份的天数差异
	birthdayRegexPattern = `^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	//手机号码规则：以1开头的11位数字，第二位数字[3-9]，剩下的九位数字可以是 0 ~ 9 之间的任意数字
	telephoneRegexPattern = `^(?:\+?86)?1[3-9]\d{9}$`
)

var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	svc                   service.UserAndService
	codeSvc               service.CodeAndService
	emailRegexExp         *regexp.Regexp
	passwordRegexExp      *regexp.Regexp
	nicknameRegex         *regexp.Regexp
	birthdayRegexPattern  *regexp.Regexp
	telephoneRegexPattern *regexp.Regexp
	ijwt.Handler
	cmd redis.Cmdable
	l   logger.LoggerV1
}

func NewUserHandler(svc service.UserAndService, codeSvc service.CodeAndService,
	jwtHandler ijwt.Handler, l logger.LoggerV1) *UserHandler {
	return &UserHandler{
		svc:                   svc,
		codeSvc:               codeSvc,
		Handler:               jwtHandler,
		l:                     l,
		emailRegexExp:         regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp:      regexp.MustCompile(passwordRegexPattern, regexp.None),
		nicknameRegex:         regexp.MustCompile(nicknameRegexPattern, regexp.None),
		birthdayRegexPattern:  regexp.MustCompile(birthdayRegexPattern, regexp.None),
		telephoneRegexPattern: regexp.MustCompile(telephoneRegexPattern, regexp.None),
	}
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type UserInfo struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var info UserInfo
	if err := ctx.Bind(&info); err != nil {
		return
	}

	//邮箱格式
	emailFlag, err := u.emailRegexExp.MatchString(info.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	if !emailFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "邮箱不正确！",
		})
		return
	}

	//密码和确认密码
	if info.Password != info.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "两次密码不相同！",
		})
		return
	}
	//密码规律
	pwdFlag, err := u.passwordRegexExp.MatchString(info.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	if !pwdFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "密码格式不正确，必须包含字母、数字、特殊字符。且长度不能小于 8 位！",
		})
		return
	}

	//存储数据...
	err = u.svc.Signup(ctx.Request.Context(), &domain.User{
		Email:    info.Email,
		Password: info.Password,
	})
	if err == service.ErrUserDuplicate {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "重复邮箱，请更换邮箱！",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "注册成功！",
	})
}

// Login 使用Session校验
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var info LoginRequest
	if err := ctx.Bind(&info); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, info.Email, info.Password)

	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "邮箱或密码不正确，请重试！",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}

	session := sessions.Default(ctx)
	session.Set("userId", user.Id)
	session.Options(sessions.Options{
		//过期时间为1分钟
		MaxAge: 60,
	})
	if err = session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功",
	})
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginJWT 使用JWT校验
func (u *UserHandler) LoginJWT(ctx *gin.Context, req LoginReq) (Result, error) {
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		//ctx.JSON(http.StatusBadRequest, Result{
		//	Code: 4,
		//	Msg:  "邮箱或密码不正确，请重试！",
		//})
		return Result{Code: 4, Msg: "邮箱或密码不正确，请重试！"}, fmt.Errorf("邮箱或密码不正确 %w！", err)
	}
	if err != nil {
		//ctx.JSON(http.StatusInternalServerError, Result{
		//	Code: 5,
		//	Msg:  "系统错误！",
		//})
		return Result{Code: 5, Msg: "系统错误！"}, nil
	}

	// 生成JWT Token
	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		//ctx.JSON(http.StatusInternalServerError, Result{
		//	Code: 5,
		//	Msg:  "系统错误！",
		//})
		return Result{Code: 5, Msg: "系统错误！"}, fmt.Errorf("生成 JWT Token 出错 %w", err)
	}

	//ctx.JSON(http.StatusOK, Result{
	//	Code: 4,
	//	Msg:  "登陆成功！",
	//})
	return Result{Msg: "登陆成功！"}, nil
}

// SendLoginSMSCode 验证码登陆
func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Request struct {
		Phone string `json:"phone"`
	}
	var req Request
	if err := ctx.Bind(&req); err != nil {
		return
	}

	phoneFlag, err := u.telephoneRegexPattern.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	if !phoneFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "手机号码不正确！",
		})
		return
	}

	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "发送成功！",
		})
	case service.ErrCodeSendTooMany:
		// 发送太频繁，需要注意是否有人蓄意破坏
		zap.L().Warn("短信发送太频繁！", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "发送太频繁，请稍后再试！",
		})
	default:
		ctx.JSON(http.StatusInternalServerError, Result{
			// 错误码系统
			Code: 5,
			Msg:  "系统错误！",
		})
	}
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// LoginSMS 校验验证码
func (u *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err == cache.ErrCodeSendTooMany {
		//ctx.JSON(http.StatusBadRequest, Result{
		//	Code: 4,
		//	Msg:  "验证次数过多！",
		//})
		return Result{Code: 4, Msg: "验证次数过多！"}, fmt.Errorf("验证次数过多 %w！", err)
	}
	if err != nil {
		//ctx.JSON(http.StatusInternalServerError, Result{
		//	Code: 5,
		//	Msg:  "系统错误！",
		//})
		//zap.L().Error("校验验证码出错！", zap.Error(err))
		// 小心敏感信息的问题，如果线上开了DEBUG级别，需要删除此段
		//zap.L().Debug("", zap.String("手机号码", req.Phone))
		return Result{Code: 5, Msg: "系统错误！"}, fmt.Errorf("校验验证码出错 %w！", err)
	}
	if !ok {
		//ctx.JSON(http.StatusBadRequest, Result{
		//	Code: 4,
		//	Msg:  "验证码有误！",
		//})
		return Result{Code: 4, Msg: "验证码有误！"}, nil
	}

	// 输入的手机号有可能是新用户
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		//ctx.JSON(http.StatusInternalServerError, Result{
		//	Code: 5,
		//	Msg:  "系统错误！",
		//})
		return Result{Code: 5, Msg: "系统错误！"}, fmt.Errorf("登陆或注册用户失败 %w！", err)
	}

	// 生成JWT Token
	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		//ctx.JSON(http.StatusInternalServerError, Result{
		//	Code: 5,
		//	Msg:  "系统错误！",
		//})
		return Result{Code: 5, Msg: "系统错误！"}, fmt.Errorf("生成 JWT Token 失败 %w！", err)
	}

	//ctx.JSON(http.StatusOK, Result{
	//	Msg:  "验证通过！",
	//})
	return Result{Msg: "验证通过！"}, nil
}

// Edit 编辑功能,允许用户补充基本个人信息
func (u *UserHandler) Edit(ctx *gin.Context) {
	type MoreInfo struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
	}

	var info MoreInfo
	if err := ctx.Bind(&info); err != nil {
		return
	}

	birthdayFlag, err := u.birthdayRegexPattern.MatchString(info.Birthday)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !birthdayFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "生日格式不正确，请以`1992-01-01`这种格式输入",
		})
		return
	}

	nicknameFlag, err := u.nicknameRegex.MatchString(info.Nickname)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !nicknameFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "昵称格式不正确，请输入2-20范围内的字符",
		})
		return
	}

	//接收login传下来的id
	session := sessions.Default(ctx)
	userId := session.Get("userId").(int64)

	if err = u.svc.Edit(ctx.Request.Context(), &domain.User{
		Id:       userId,
		Birthday: info.Birthday,
		Nickname: info.Nickname,
	}); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "更新信息失败，请检查格式",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "更新个人信息成功",
	})
}

// EditJWT 编辑功能,允许用户补充基本个人信息
func (u *UserHandler) EditJWT(ctx *gin.Context) {
	type MoreInfo struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
	}

	var info MoreInfo
	if err := ctx.Bind(&info); err != nil {
		return
	}

	birthdayFlag, err := u.birthdayRegexPattern.MatchString(info.Birthday)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	if !birthdayFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "生日格式不正确，请以`1992-01-01`这种格式输入！",
		})
		return
	}

	nicknameFlag, err := u.nicknameRegex.MatchString(info.Nickname)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误！",
		})
		return
	}
	if !nicknameFlag {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "昵称格式不正确，请输入2-20范围内的字符！",
		})
		return
	}

	//接收login传下来的id
	//接收login传下来的claims
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 3,
			Msg:  "未授权!",
		})
		return
	}

	if err = u.svc.Edit(ctx.Request.Context(), &domain.User{
		Id:       claims.UserId,
		Birthday: info.Birthday,
		Nickname: info.Nickname,
	}); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "更新信息失败，请检查格式！",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "更新个人信息成功！",
	})
}

// Profile 查看用户详情
func (u *UserHandler) Profile(ctx *gin.Context) {
	session := sessions.Default(ctx)
	userId := session.Get("userId").(int64)

	user, err := u.svc.Profile(ctx, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"nickname":     user.Nickname,
		"email":        user.Email,
		"phone number": user.Phone,
		"birthday":     user.Birthday,
		"create_at":    user.Ctime,
		"update_at":    user.Utime,
	})
}

// ProfileJWT 查看用户详情
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	user, err := u.svc.Profile(ctx, claims.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"nickname":     user.Nickname,
		"email":        user.Email,
		"phone number": user.Phone,
		"birthday":     user.Birthday,
		"create_at":    user.Ctime,
		"update_at":    user.Utime,
	})
}

// Exit 退出功能
func (u *UserHandler) Exit(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	if err := sess.Save(); err != nil {
		ctx.String(http.StatusInternalServerError, "系统故障")
		return
	}
	ctx.String(http.StatusOK, "退出登陆成功！")
}

func (u *UserHandler) ExitJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "退出登陆失败！",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "退出登陆成功！",
	})
}

// RefreshToken 可以同时刷新长短 token，用 redis 来记录是否有效，即 refresh_token 是一次性的
// 参考登陆校验部分，比较 User-Agent 来增强安全性
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 只有从这里拿出来的是 refresh_token，其他地方拿到的都是 access_token
	rfTokenString := u.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	acToken, err := jwt.ParseWithClaims(rfTokenString, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RefreshTokenKey, nil
	})
	if err != nil || !acToken.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 确认使用该 token 的用户还没退出登录
	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil || !acToken.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 创建新的 access_token
	err = u.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		zap.L().Error("设置 JWT token 时出现异常！", zap.Error(err),
			zap.String("method", "UserHandler:RefreshToken"))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}
