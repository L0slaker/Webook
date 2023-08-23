package web

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

const (
	//邮箱规则
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	//密码规则：至少1个大小写字母、一个数字、一个特殊符号，长度至少为8个字符
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	//昵称规则：长度为2到20个字符，可以是字母（大小写）、数字或下划线。
	nicknameRegexPattern = `^[A-Za-z0-9_]{2,20}$`
	//生日规则：以19或20开头的四位数字年份、'-'分隔的两位数字的月份，范围从01到12、'-'分隔的两位数字的日期，范围从01到31,考虑了不同月份的天数差异
	birthdayRegexPattern = `^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
)

type UserHandler struct {
	svc                  *service.UserService
	emailRegexExp        *regexp.Regexp
	passwordRegexExp     *regexp.Regexp
	nicknameRegex        *regexp.Regexp
	birthdayRegexPattern *regexp.Regexp
	jwtKey               string
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc:                  svc,
		emailRegexExp:        regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp:     regexp.MustCompile(passwordRegexPattern, regexp.None),
		nicknameRegex:        regexp.MustCompile(nicknameRegexPattern, regexp.None),
		birthdayRegexPattern: regexp.MustCompile(birthdayRegexPattern, regexp.None),
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
		ctx.String(http.StatusInternalServerError, "系统错误")
	}
	if !emailFlag {
		ctx.String(http.StatusBadRequest, "邮箱不正确")
		return
	}

	//密码和确认密码
	if info.Password != info.ConfirmPassword {
		ctx.String(http.StatusBadRequest, "两次密码不相同")
		return
	}
	//密码规律
	pwdFlag, err := u.passwordRegexExp.MatchString(info.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
	}
	if !pwdFlag {
		ctx.String(http.StatusBadRequest, "密码格式不正确，必须包含字母、数字、特殊字符。且长度不能小于 8 位")
		return
	}

	//存储数据...
	if err = u.svc.Signup(ctx.Request.Context(), &domain.User{
		Email:    info.Email,
		Password: info.Password,
	}); err != nil {
		ctx.String(http.StatusBadRequest, "重复邮箱，请更换邮箱！")
		return
	}
	ctx.String(http.StatusOK, "注册成功！")
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
		ctx.String(http.StatusBadRequest, "邮箱或密码不正确，请重试")
		return
	}

	session := sessions.Default(ctx)
	session.Set("userId", user.Id)
	session.Options(sessions.Options{
		//过期时间为1分钟
		MaxAge: 60,
	})
	if err = session.Save(); err != nil {
		ctx.String(http.StatusInternalServerError, "服务器异常")
		return
	}
	ctx.String(http.StatusOK, "登陆成功")
}

// LoginJWT 使用JWT校验
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
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
		ctx.String(http.StatusBadRequest, "邮箱或密码不正确，请重试")
		return
	}

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenString)

	ctx.String(http.StatusOK, "登陆成功")
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
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !birthdayFlag {
		ctx.String(http.StatusBadRequest, "生日格式不正确，请以`1992-01-01`这种格式输入")
		return
	}

	nicknameFlag, err := u.nicknameRegex.MatchString(info.Nickname)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !nicknameFlag {
		ctx.String(http.StatusBadRequest, "昵称格式不正确，请输入2-20范围内的字符")
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
		ctx.String(http.StatusBadRequest, "更新信息失败，请检查格式")
		return
	}

	ctx.String(http.StatusOK, "更新个人信息成功")
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
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !birthdayFlag {
		ctx.String(http.StatusBadRequest, "生日格式不正确，请以`1992-01-01`这种格式输入")
		return
	}

	nicknameFlag, err := u.nicknameRegex.MatchString(info.Nickname)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	if !nicknameFlag {
		ctx.String(http.StatusBadRequest, "昵称格式不正确，请输入2-20范围内的字符")
		return
	}

	//接收login传下来的id
	//接收login传下来的claims
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusUnauthorized, "系统错误")
		return
	}

	if err = u.svc.Edit(ctx.Request.Context(), &domain.User{
		Id:       claims.UserId,
		Birthday: info.Birthday,
		Nickname: info.Nickname,
	}); err != nil {
		ctx.String(http.StatusBadRequest, "更新信息失败，请检查格式")
		return
	}

	ctx.String(http.StatusOK, "更新个人信息成功")
}

// Profile 查看用户详情
func (u *UserHandler) Profile(ctx *gin.Context) {
	type LoginEmail struct {
		Email    string
		Nickname string
		Birthday string
	}

	var info LoginEmail
	if err := ctx.Bind(&info); err != nil {
		return
	}

	session := sessions.Default(ctx)
	userId := session.Get("userId").(int64)

	user, err := u.svc.Profile(ctx, userId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user info": LoginEmail{
			Email:    user.Email,
			Nickname: user.Nickname,
			Birthday: user.Birthday,
		},
		"create_at": user.CreateTime,
		"update_at": user.UpdateTime,
	})
}

// ProfileJWT 查看用户详情
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type LoginEmail struct {
		Email    string
		Nickname string
		Birthday string
	}

	var info LoginEmail
	if err := ctx.Bind(&info); err != nil {
		return
	}

	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, claims.UserId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user info": LoginEmail{
			Email:    user.Email,
			Nickname: user.Nickname,
			Birthday: user.Birthday,
		},
		"create_at": user.CreateTime,
		"update_at": user.UpdateTime,
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

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	UserAgent string
}
