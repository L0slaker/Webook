package sms

import "context"

type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
	//SendV1(ctx context.Context, tplId string, args []NameArg, numbers ...string) error
}

type NameArg struct {
	Name string
	Val  string
}
