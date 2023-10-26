package ratelimit

import (
	"Prove/webook/internal/service/sms"
	smsmocks "Prove/webook/internal/service/sms/mocks"
	"Prove/webook/pkg/ratelimit"
	limitmocks "Prove/webook/pkg/ratelimit/mocks"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRatelimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) (sms.Service, ratelimit.Limiter)
		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return svc, limit
			},
			wantErr: nil,
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return svc, limit
			},
			wantErr: ErrLimited,
		},
		{
			name: "限流器异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("限流器异常"))
				return svc, limit
			},
			wantErr: fmt.Errorf("短信服务判断是否限流出现问题：%w", errors.New("限流器异常")),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, limiter := tc.mock(ctrl)
			limitSvc := NewRatelimitSMSService(svc, limiter)
			err := limitSvc.Send(context.Background(), "my_template",
				[]string{"123456"}, "13376807963")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
