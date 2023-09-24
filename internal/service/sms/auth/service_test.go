package auth

import (
	"Prove/webook/internal/service/sms"
	smsmocks "Prove/webook/internal/service/sms/mocks"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestSMSService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) (sms.Service, []byte)
		wantErr error
	}{
		{
			name: "解析失败",
		},
		{
			name: "token不合法",
		},
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, []byte) {
				svc := smsmocks.NewMockService(ctrl)
				key := []byte("say_goodbye")
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return svc, key
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewSMSService(tc.mock(ctrl))
			err := svc.Send(context.Background(), "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJUcGwiOiJteV90ZW1wbGF0ZSJ9.-UM-WI4yifIx5Tvj9mbGnnEQpcwlRwxgXdRI0wFzipIKIG9wcULqz9xLp8lN_zc11sgTWQ2tjY0ZdPZ84ZoskQ",
				[]string{"123456"}, "13376807963")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_GenerateJWTToken(t *testing.T) {
	claims := Claims{
		Tpl: "my_template",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	key := []byte("say_goodbye")
	tokenString, _ := token.SignedString(key)
	fmt.Println(tokenString)
}
