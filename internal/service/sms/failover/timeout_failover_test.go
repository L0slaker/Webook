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

func TestTimeoutFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		idx       int32
		cnt       int32
		threshold int32
		mock      func(*gomock.Controller) []sms.Service
		wantIdx   int32
		wantCnt   int32
		wantErr   error
	}{
		{
			name: "超时，但没连续超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc1}
			},
			threshold: 3,
			wantIdx:   0,
			wantCnt:   1,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "触发了切换，切换之后成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			wantIdx:   1,
			wantCnt:   0,
			wantErr:   nil,
		},
		{
			name: "触发了切换，切换之后失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败"))
				return []sms.Service{svc1, svc2}
			},
			threshold: 3,
			cnt:       3,
			wantCnt:   0,
			wantIdx:   1,
			wantErr:   errors.New("发送失败"),
		},
		{
			name: "触发了切换，切换之后依旧超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc1, svc2}
			},
			threshold: 3,
			cnt:       3,
			wantCnt:   1,
			wantIdx:   1,
			wantErr:   context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewTimeoutFailoverSMSService(tc.mock(ctrl), tc.threshold)
			svc.idx = tc.idx
			svc.cnt = tc.cnt

			err := svc.Send(context.Background(), "my_template",
				[]string{"134889"}, "15380007964")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantIdx, svc.idx)
			assert.Equal(t, tc.wantCnt, svc.cnt)
		})
	}
}
