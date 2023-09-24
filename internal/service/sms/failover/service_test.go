package failover

import (
	"Prove/webook/internal/service/sms"
	smsmocks "Prove/webook/internal/service/sms/mocks"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) []sms.Service
		wantErr error
	}{
		{
			name: "一次成功！",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return []sms.Service{svc1}
			},
			wantErr: nil,
		},
		{
			name: "重试成功！",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败！"))
				svc2 := smsmocks.NewMockService(ctrl)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return []sms.Service{svc1, svc2}
			},
			wantErr: nil,
		},
		{
			name: "重试最终失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败！"))
				svc2 := smsmocks.NewMockService(ctrl)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败！"))
				svc3 := smsmocks.NewMockService(ctrl)
				svc3.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败！"))
				return []sms.Service{svc1, svc2, svc3}
			},
			wantErr: errors.New("发送失败，所有的服务商都尝试过了"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewFailoverSMSService(tc.mock(ctrl))
			err := svc.Send(context.Background(), "my_template",
				[]string{"113679"}, "15259907833")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
