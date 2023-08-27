package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁！")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数过多！")
	ErrUnknownForCode         = errors.New("未知的错误！")
)

// 编译器将lua/set_code.lua文件的内容嵌入到了luaSetCode字符串变量中
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	//Eval 执行脚本的方法；由于脚本需要int类型的返回值，所以调用了Int()
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		//没有问题
		return nil
	case -1:
		//发送频繁
		return ErrCodeSendTooMany
	//case -2: 本质上和default一致
	default:
		//系统错误，有人误操作
		return errors.New("系统错误！")
	}
}

func (c *CodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForCode
	}
}

func (c *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
