package auth

import (
	"Prove/webook/internal/service/sms"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key []byte
}

func NewSMSService(svc sms.Service, key []byte) sms.Service {
	return &SMSService{
		svc: svc,
		key: key,
	}
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var c Claims
	// 如果这里能够解析成功，就可以说明是对应的业务方
	token, err := jwt.ParseWithClaims(biz, &c, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token 不合法")
	}

	return s.svc.Send(ctx, biz, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
