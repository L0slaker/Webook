package web

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/service"
	svcmocks "Prove/webook/internal/service/mocks"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
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

func TestNil(t *testing.T) {
	testTypeAssert(nil)
}

func testTypeAssert(c any) {
	_, ok := c.(*UserClaims)
	println(ok)
}

// Handler测试的主要难点
// 1.构造HTTP请求
// 2.验证HTTP响应
func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserAndService
		body     string
		wantCode int
		wantBody Result
	}{
		{
			name: "注册成功！",
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(nil)
				return userSvc
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
		{
			name: "绑定信息错误！",
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				return userSvc
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
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				// 在检测格式时就被return，没有进入Signup方法中
				//userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
				//	Email:    "l0slakers",
				//	Password: "Abcd#1234",
				//})
				return userSvc
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
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				//userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
				//	Email:    "l0slakers@gmail.com",
				//	Password: "Abcd#1234",
				//})
				return userSvc
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
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				return userSvc
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
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(service.ErrUserDuplicate)
				return userSvc
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
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), &domain.User{
					Email:    "l0slakers@gmail.com",
					Password: "Abcd#1234",
				}).Return(errors.New("any error"))
				return userSvc
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
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
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
		name     string
		mock     func(*gomock.Controller) service.UserAndService
		body     string
		wantCode int
		wantBody Result
	}{
		{
			name: "解析数据失败",
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				return userSvc
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
		},
		{
			name: "系统错误",
		},
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) service.UserAndService {
				userSvc := svcmocks.NewMockUserAndService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(),
					"l0slakers@gmail.com",
					"Abcd#1234").Return(&domain.User{}, nil)
				return userSvc
			},
			body: `
{
	"email": "l0slakers@gmail.com",
	"password": "Abcd#1234"
}
`,
			wantCode: http.StatusOK,
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
			h := NewUserHandler(tc.mock(ctrl), nil)
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

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, respBody)
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
