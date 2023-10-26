package integration

import (
	"Prove/webook/internal/integration/startup"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/internal/web"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/ioc"
	"Prove/webook/pkg/logger"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// UserTestSuite 测试套件
type UserTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

// SetupSuite 初始化测试内容
func (u *UserTestSuite) SetupSuite() {
	u.server = gin.Default()
	u.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			UserId: 123,
		})
	})
	u.db = startup.InitTestDB()

}

func TestUserHandler_e2e_SignUp(t *testing.T) {
	server := startup.InitWebServer()
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	db := ioc.InitDB(logger.NewZapLogger(l))
	now := time.Now()
	testCases := []struct {
		name     string
		body     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		wantCode int
		wantBody web.Result
	}{
		{
			name: "bind 失败！",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234",
	"confirmPassword": "
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不正确！",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			body: `
{
	"email": "l0slakers",
	"password": "Abcd#1234",
	"confirmPassword": "Abcd#1234"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "邮箱不正确！",
			},
		},
		{
			name: "两次密码不相同！",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234",
	"confirmPassword": "Ac#123456"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "两次密码不相同！",
			},
		},
		{
			name: "密码格式不正确！",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "56781234",
	"confirmPassword": "56781234"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "密码格式不正确，必须包含字母、数字、特殊字符。且长度不能小于 8 位！",
			},
		},
		{
			name: "邮箱冲突！",
			before: func(t *testing.T) {
				u := dao.User{
					Email: sql.NullString{
						String: "l0slakers@gmail.com",
						Valid:  true,
					},
					Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					CreateTime: now.UnixMilli(),
					UpdateTime: now.UnixMilli(),
				}
				db.Create(&u)
			},
			after: func(t *testing.T) {
				var u dao.User
				d := db.Where("email = ?", "l0slakers@gmail.com").First(&u)
				d.Delete(&u)
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234",
	"confirmPassword": "Abcd#1234"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "重复邮箱，请更换邮箱！",
			},
		},
		{
			name: "注册成功！",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var u dao.User
				d := db.Where("email = ?", "l0slakers@gmail.com").First(&u)
				assert.NotEmpty(t, u.Id)
				assert.NotEmpty(t, u.Email)
				assert.NotEmpty(t, u.Password)
				assert.NotEmpty(t, u.CreateTime)
				assert.NotEmpty(t, u.UpdateTime)
				d.Delete(&u)
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234",
	"confirmPassword": "Abcd#1234"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "注册成功！",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.body)))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			var respBody web.Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			tc.after(t)
		})
	}
}

func TestUserHandler_e2e_LoginJWT(t *testing.T) {
	server := startup.InitWebServer()
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	db := ioc.InitDB(logger.NewZapLogger(l))
	now := time.Now()
	testCases := []struct {
		name     string
		body     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		wantCode int
		wantBody web.Result
	}{
		{
			name: "bind 失败！",
			before: func(t *testing.T) {
				u := dao.User{
					Email: sql.NullString{
						String: "l0slakers@gmail.com",
						Valid:  true,
					},
					Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					CreateTime: now.UnixMilli(),
					UpdateTime: now.UnixMilli(),
				}
				db.Create(&u)
			},
			after: func(t *testing.T) {
				var u dao.User
				d := db.Where("email = ?", "l0slakers@gmail.com").First(&u)
				d.Delete(&u)
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱或密码不正确！",
			before: func(t *testing.T) {
				u := dao.User{
					Email: sql.NullString{
						String: "l0slakers@gmail.com",
						Valid:  true,
					},
					// 该密码是 Abcd#1234 生成的
					Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					CreateTime: now.UnixMilli(),
					UpdateTime: now.UnixMilli(),
				}
				db.Create(&u)
			},
			after: func(t *testing.T) {
				var u dao.User
				d := db.Where("email = ?", "l0slakers@gmail.com").First(&u)
				err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("1234Abcd#"))
				ErrInvalidUserOrPassword := errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password")
				require.Equal(t, ErrInvalidUserOrPassword, err)
				d.Delete(&u)
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "1234Abcd#"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "邮箱或密码不正确，请重试！",
			},
		},
		{
			name: "登陆成功！",
			before: func(t *testing.T) {
				u := dao.User{
					Email: sql.NullString{
						String: "l0slakers@gmail.com",
						Valid:  true,
					},
					Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					CreateTime: now.UnixMilli(),
					UpdateTime: now.UnixMilli(),
				}
				db.Create(&u)
			},
			after: func(t *testing.T) {
				var u dao.User
				d := db.Where("email = ?", "l0slakers@gmail.com").First(&u)
				err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("Abcd#1234"))
				require.Equal(t, nil, err)
				d.Delete(&u)
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "登陆成功！",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			var respBody web.Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			tc.after(t)
		})
	}
}

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string
		body string
		// 准备数据
		before func(t *testing.T)
		// 验证并清理数据
		after    func(t *testing.T)
		wantCode int
		wantBody web.Result
	}{
		{
			name: "bind 失败！",
			before: func(t *testing.T) {
				// redis 里面没数据
			},
			after: func(t *testing.T) {

			},
			body: `
{
	"phone":
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "手机号码不正确！",
			before: func(t *testing.T) {
				// redis 里面没数据
			},
			after: func(t *testing.T) {

			},
			body: `
{
	"phone":""
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "手机号码不正确！",
			},
		},
		{
			name: "发送成功！",
			before: func(t *testing.T) {
				// redis 里面没数据
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				val, err := rdb.GetDel(ctx, "phone_code:login:13509516670").Result()
				cancel()
				assert.NoError(t, err)
				// 6 位验证码
				assert.True(t, len(val) == 6)
			},
			body: `
{
	"phone":"13509516670"
}
`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "发送成功！",
			},
		},
		{
			name: "发送太频繁！",
			before: func(t *testing.T) {
				// 这个手机号码已经发送了验证码
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				_, err := rdb.Set(ctx, "phone_code:login:13509516670", "123456",
					time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				val, err := rdb.GetDel(ctx, "phone_code:login:13509516670").Result()
				cancel()
				assert.NoError(t, err)
				// 验证码还是1234556，没有被覆盖
				assert.Equal(t, "123456", val)
			},
			body: `
{
	"phone":"13509516670"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: web.Result{
				Code: 4,
				Msg:  "发送太频繁，请稍后再试！",
			},
		},
		{
			name: "系统错误！",
			before: func(t *testing.T) {
				// 这个手机号码，已经有了一个验证码，但没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				_, err := rdb.Set(ctx, "phone_code:login:13509516670", "123456",
					0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				val, err := rdb.GetDel(ctx, "phone_code:login:13509516670").Result()
				cancel()
				assert.NoError(t, err)
				// 6 位验证码
				assert.Equal(t, "123456", val)
			},
			body: `
{
	"phone":"13509516670"
}
`,
			wantCode: http.StatusInternalServerError,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/send/code", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			var respBody web.Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			tc.after(t)
		})
	}
}

//func TestUserHandler_e2e_EditJWT(t *testing.T) {
//	server := InitWebServer()
//	db := ioc.InitDB()
//	//now := time.Now()
//	testCases := []struct {
//		name     string
//		body     string
//		before   func(t *testing.T)
//		after    func(t *testing.T)
//		wantCode int
//		wantBody web.Result
//	}{
//		{
//			name: "更新个人信息成功！",
//			before: func(t *testing.T) {
//				u := dao.User{
//					Id: 123,
//				}
//				db.Create(&u)
//			},
//			after: func(t *testing.T) {
//				var u dao.User
//				ctx := &gin.Context{Request: &http.Request{}}
//				server.HandleContext(ctx)
//				claim := ctx.Value("claim")
//				c := claim.(*web.UserClaims)
//
//				d := db.Where("id = ?", c.UserId).First(&u)
//				d.UpdateColumn("nickname", "Kobe")
//				d.UpdateColumn("birthday", "2000-12-22")
//				d.Delete(&u)
//			},
//			body: `
//{
//	"nickname": "Kobe",
//	"birthday":"2000-12-22"
//}
//`,
//			wantCode: http.StatusOK,
//			wantBody: web.Result{
//				Code: 4,
//				Msg:  "更新个人信息成功！",
//			},
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			tc.before(t)
//			req, err := http.NewRequest(http.MethodPost, "/users/edit", bytes.NewBuffer([]byte(tc.body)))
//			require.NoError(t, err)
//			req.Header.Set("Content-Type", "application/json")
//			//req.Header.Set("Authorization", "x-jwt-token")
//
//			resp := httptest.NewRecorder()
//
//			// 中间件校验的过程
//			server.Use(func(ctx *gin.Context) {
//				ctx.Set("claim", 123)
//			})
//
//			server.ServeHTTP(resp, req)
//
//			var respBody web.Result
//			err = json.NewDecoder(resp.Body).Decode(&respBody)
//			if err != nil {
//				assert.Equal(t, errors.New("EOF"), err)
//			}
//
//			assert.Equal(t, tc.wantCode, resp.Code)
//			assert.Equal(t, tc.wantBody, respBody)
//			tc.after(t)
//		})
//	}
//}

func TestUserHandler_e2e_ProfileJWT(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestUserHandler_e2e_LoginSMS(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
