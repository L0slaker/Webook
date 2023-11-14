package errs

// 公共模块
const (
	UserUnauthorizedError = 400001 // 用户未登陆
)

// 账号密码登陆模块
const (
	UserInvalidInput        = 401001 // 用户输入错误
	UserInvalidOrPassword   = 401002 // 用户不存在或密码错误
	UserDuplicateEmail      = 401002 // 重复邮箱
	UserInternalServerError = 501001
)

// 短信登陆模块
const (
	CodeInvalidInput              = 402001 // 用户输入错误
	CodeSendTooMany               = 402002 // 验证码发送太频繁
	CodeVerifyTooMany             = 402003 // 验证码超出次数
	CodeVerifyError               = 402004 // 验证码错误
	CodeSendInternalServerError   = 502001
	CodeVerifyInternalServerError = 502002
)

// 微信登陆模块
const (
	WechatLoginFailed         = 403001 // 登陆失败
	WechatInternalServerError = 503001
)

// 文章模块
const (
	ArticleInvalidInput        = 404001 // 用户输入错误
	ArticleInternalServerError = 504001
)
