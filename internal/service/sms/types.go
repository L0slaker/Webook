package sms

import "context"

type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
	//Send(ctx context.Context, tplId string, args []string, numbers ...string) error
	//SendV1(ctx context.Context, tplId string, args []NameArg, numbers ...string) error
}

type NameArg struct {
	Name string
	Val  string
}
