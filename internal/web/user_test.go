package web

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/service"
	svcmocks "Prove/webook/internal/service/mocks"
	ijwt "Prove/webook/internal/web/jwt"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := "l0slakers@gmail.com"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

// Handler测试的主要难点
// 1.构造HTTP请求
// 2.验证HTTP响应
func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler)
		body     string
		wantCode int
		wantBody Result
	}{
		{
			name: "绑定信息错误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers@gmail.com",
			"password": "Abcd#1234"
		`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱不正确！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				// 在检测格式时就被return，没有进入Signup方法中
				//userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
				//	Email:    "l0slakers",
				//	Password: "Abcd#1234",
				//})
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers",
			"password": "Abcd#1234",
			"confirmPassword": "Abcd#1234"
		}
		`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "邮箱不正确！",
			},
		},
		{
			name: "两次输入密码不一致！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				//userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
				//	Email:    "l0slakers@gmail.com",
				//	Password: "Abcd#1234",
				//})
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers@gmail.com",
			"password": "Abcd#12345678",
			"confirmPassword": "Abcd#1234"
		}
		`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "两次密码不相同！",
			},
		},
		{
			name: "密码格式不正确！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers@gmail.com",
			"password": "123456",
			"confirmPassword": "123456"
		}
		`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "密码格式不正确，必须包含字母、数字、特殊字符。且长度不能小于 8 位！",
			},
		},
		{
			name: "重复邮箱！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(service.ErrUserDuplicate)
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers@gmail.com",
			"password": "Abcd#1234",
			"confirmPassword": "Abcd#1234"
		}
		`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "重复邮箱，请更换邮箱！",
			},
		},
		{
			name: "系统错误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(errors.New("any error"))
				return userSvc, codeSvc, nil
			},
			body: `
		{
			"email": "l0slakers@gmail.com",
			"password": "Abcd#1234",
			"confirmPassword": "Abcd#1234"
		}
		`,
			wantCode: http.StatusInternalServerError,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
		{
			name: "注册成功！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(nil)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234",
	"confirmPassword": "Abcd#1234"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "注册成功！",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			// 设置请求头
			req.Header.Set("Content-Type", "application/json")
			// http请求的记录
			resp := httptest.NewRecorder()

			// HTTP 请求进入 GIN 框架的入口
			// 调用此方法时，Gin 会处理这个请求，将响应写回 resp 里
			r.ServeHTTP(resp, req)
			//
			var respBody Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
		})
	}
}

// 注意 RegisterRoutes()、loginHandler() 中注册的路由是Session的实现还是JWT的实现
func TestUserHandler_LoginJWT(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler)
		body       string
		wantUserId int64
		wantCode   int
		wantBody   Result
	}{
		{
			name: "解析数据失败",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱或密码不正确",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(),
					"l0slakers@gmail.com",
					"Abcd#1234").Return(&domain.User{}, service.ErrInvalidUserOrPassword)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "邮箱或密码不正确，请重试！",
			},
		},
		{
			name: "系统错误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(),
					"l0slakers@gmail.com",
					"Abcd#1234").Return(&domain.User{}, errors.New("any error"))
				return userSvc, codeSvc, nil
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
}
`,
			wantCode: http.StatusInternalServerError,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
		{
			name: "登陆成功!",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(),
					"l0slakers@gmail.com",
					"Abcd#1234").Return(&domain.User{
					Id: 123,
				}, nil)
				handler := ijwt.NewRedisJWT(nil)
				return userSvc, codeSvc, handler
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
}
`,
			wantUserId: 123,
			wantCode:   http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "登陆成功！",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			var respBody Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			tokenString := resp.Header().Get("x-jwt-token")
			claims := &ijwt.UserClaims{}
			_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), nil
			})
			if err != nil {
				t.Log(err)
				fmt.Println("解析token 失败")
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			assert.Equal(t, tc.wantUserId, claims.UserId)
		})
	}
}

func TestUserHandler_EditJWT(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler)
		//reqBuilder func(t *testing.T) *http.Request
		body       string
		wantUserId int64
		wantCode   int
		wantBody   Result
	}{
		{
			name: "更新成功！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&domain.User{
						Id: 123,
					}, nil)
				handler := ijwt.NewRedisJWT(nil)
				userSvc.EXPECT().Edit(gomock.Any(), &domain.User{
					Id:       123,
					Birthday: "2000-12-14",
					Nickname: "Lakers",
				}).Return(nil)
				return userSvc, codeSvc, handler
			},
			body: `
{
	"nickname":"Lakers",
	"birthday":"2000-12-14"
}
`,
			wantUserId: 123,
			wantCode:   http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "更新个人信息成功",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/users/edit", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "x-jwt-token")

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			tokenString := resp.Header().Get("x-jwt-token")
			claims := &ijwt.UserClaims{
				UserId: 123,
			}
			_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), nil
			})
			if err != nil {
				t.Log(err)
				t.Fatal("解析token 失败")
				return
			}

			var respBody Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
				return
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			assert.Equal(t, tc.wantUserId, claims.UserId)
		})
	}
}

func TestUserHandler_SendLoginSMSCode(t *testing.T) {
	testCases := []struct {
		name     string
		body     string
		mock     func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler)
		wantCode int
		wantBody Result
	}{
		{
			name: "bind 失败！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"Phone": "13500997890"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "手机号码格式不正确！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"Phone": "11416"
}`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "手机号码不正确！",
			},
		},
		{
			name: "发送成功！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"Phone": "13500997890"
}`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "发送成功！",
			},
		},
		{
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any()).Return(service.ErrCodeSendTooMany)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"Phone": "13500997890"
}`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "发送太频繁，请稍后再试！",
			},
		},
		{
			name: "系统错误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("系统错误！"))
				return userSvc, codeSvc, nil
			},
			body: `
{
	"Phone": "13500997890"
}`,
			wantCode: http.StatusInternalServerError,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/send/code", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(r)

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			var respBody Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(*gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler)
		body       string
		wantUserId int64
		wantCode   int
		wantBody   Result
	}{
		{
			name: "bind 失败！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"phone":"13370898966",
	"code":"123456"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "验证码有误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(false, nil)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"phone":"13370898966",
	"code":"123456"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "验证码有误！",
			},
		},
		{
			name: "验证次数过多！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(false, cache.ErrCodeSendTooMany)
				return userSvc, codeSvc, nil
			},
			body: `
{
	"phone":"13370898966",
	"code":"123456"
}
`,
			wantCode: http.StatusBadRequest,
			wantBody: Result{
				Code: 4,
				Msg:  "验证次数过多！",
			},
		},
		{
			name: "系统错误！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), gomock.Any()).
					Return(&domain.User{}, errors.New("系统错误！"))
				return userSvc, codeSvc, nil
			},
			body: `
{
	"phone":"13370898966",
	"code":"123456"
}
`,
			wantCode: http.StatusInternalServerError,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
		{
			name: "验证通过！",
			mock: func(ctrl *gomock.Controller) (service.UserAndService, service.CodeAndService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				codeSvc := svcmocks.NewMockCodeAndService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), gomock.Any()).
					Return(&domain.User{
						Id: 123,
					}, nil)
				handler := ijwt.NewRedisJWT(nil)
				return userSvc, codeSvc, handler
			},
			body: `
{
	"phone":"13370898966",
	"code":"123456"
}
`,
			wantCode:   http.StatusOK,
			wantUserId: 123,
			wantBody: Result{
				Code: 4,
				Msg:  "验证通过！",
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			var respBody Result
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			if err != nil {
				assert.Equal(t, errors.New("EOF"), err)
			}

			tokenString := resp.Header().Get("x-jwt-token")
			claims := &ijwt.UserClaims{}
			_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("OAFXibGNCqeU49DiXzCADjs9up9d7bJz"), nil
			})
			if err != nil {
				t.Log(err)
				fmt.Println("解析token 失败")
			}

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
			assert.Equal(t, tc.wantUserId, claims.UserId)
		})
	}
}

func TestMock(t *testing.T) {
	//mock的使用：
	//1.初始化控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	//2.创建模拟的对象
	usersvc := svcmocks.NewMockUserAndService(ctrl)
	//3.设计模拟调用
	//	3.1先调用 EXPECT
	//	3.2调用同名方法，传入模拟的条件
	//	3.3指定返回值

	usersvc.EXPECT().Signup(gomock.Any(), &domain.User{
		Email:    "l0slakers@gmail.com",
		Password: "Abcd#1234",
	}).Return(errors.New("mock error"))

	err := usersvc.Signup(context.Background(), &domain.User{
		Email:    "l0slakers@gmail.com",
		Password: "Abcd#1234",
	})
	t.Log(err)
}
