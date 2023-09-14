package cache

import (
	"Prove/webook/internal/repository/cache/redismocks"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name:  "发送成功！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			wantErr: nil,
		},
		{
			name:  "redis 错误！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("mock redis 错误！"))
				//res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			wantErr: errors.New("mock redis 错误！"),
		},
		{
			name:  "发送频繁！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			wantErr: ErrCodeSendTooMany,
		},
		{
			name:  "系统错误！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-2))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			wantErr: errors.New("系统错误！"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			codeCache := NewRedisCodeCache(tc.mock(ctrl))
			err := codeCache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestRedisCodeCache_Verify(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		isSend  bool
		wantErr error
	}{
		{
			name:  "验证通过！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			isSend:  true,
			wantErr: nil,
		},
		{
			name:  "redis 错误！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("mock redis 错误！"))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			isSend:  false,
			wantErr: errors.New("mock redis 错误！"),
		},
		{
			name:  "发送频繁！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			isSend:  false,
			wantErr: ErrCodeVerifyTooManyTimes,
		},
		{
			name:  "系统错误！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-2))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			isSend:  false,
			wantErr: nil,
		},
		{
			name:  "未知错误！",
			ctx:   context.Background(),
			biz:   "login",
			phone: "13378615998",
			code:  "114567",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-3))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode,
					[]string{"phone_code:login:13378615998"},
					[]any{"114567"}).Return(res)
				return cmd
			},
			isSend:  false,
			wantErr: ErrUnknownForCode,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			codeCache := NewRedisCodeCache(tc.mock(ctrl))
			isSend, err := codeCache.Verify(tc.ctx, tc.biz, tc.phone, tc.code)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.isSend, isSend)
		})
	}
}
