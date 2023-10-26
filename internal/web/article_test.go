package web

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/service"
	svcmocks "Prove/webook/internal/service/mocks"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/pkg/logger"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Edit(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		body     string
		wantCode int
		wantRes  Result
	}{
		{
			name: "bind 失败！",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "保存帖子失败！",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Save(gomock.Any(), &domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), errors.New("save error"))
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容"
}
`,
			wantCode: http.StatusInternalServerError,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
		{
			name: "保存成功！",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Save(gomock.Any(), &domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Msg:  "保存成功！",
				Data: float64(1),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			r.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.UserClaims{
					UserId: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var webRes Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)

			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		body     string
		wantCode int
		wantRes  Result
	}{
		{
			name: "bind 失败！",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "发表失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Publish(gomock.Any(), &domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish error"))
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容"
}
`,
			wantCode: http.StatusInternalServerError,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误！",
			},
		},
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Publish(gomock.Any(), &domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return artSvc
			},
			body: `
{
	"title":"我的标题",
	"content":"我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Msg:  "发布成功！",
				Data: float64(1),
			},
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := gin.Default()
			// 模拟登录态
			r.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.UserClaims{
					UserId: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
			h.RegisterRoutes(r)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var webRes Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)

			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
