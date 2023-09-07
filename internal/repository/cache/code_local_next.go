package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LocalCodeCacheV2 struct {
	store *sync.Map // 存储验证码的本地缓存
}

type CodeInfoV2 struct {
	Code         string
	Expiration   time.Time
	Verification int
}

func NewLocalCodeCacheV2(store *sync.Map) CodeCache {
	return &LocalCodeCacheV2{
		store: store,
	}
}

func (c *LocalCodeCacheV2) Set(ctx context.Context, biz, phone, code string) error {
	cacheKey := c.key(biz, phone)

	val, ok := c.store.Load(cacheKey)
	if ok {
		info := val.(CodeInfoV2)
		// 检查是否发送过频繁
		elapsed := time.Since(info.Expiration)
		if elapsed < 60*time.Second {
			return ErrCodeSendTooMany
		}

		// 删除过期的验证码
		if elapsed >= 600*time.Second {
			c.store.Delete(cacheKey)
			return ErrCodeSendExpired
		}
	}

	// 将验证码存入本地缓存
	info := CodeInfoV2{
		Code:         code,
		Expiration:   time.Now(),
		Verification: 0,
	}
	c.store.Store(cacheKey, info)
	return nil
}

func (c *LocalCodeCacheV2) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	cacheKey := c.key(biz, phone)

	// 获取存储的验证码
	val, ok := c.store.Load(cacheKey)
	if !ok {
		return false, nil
	}

	info := val.(CodeInfoV2)
	// 检查验证次数
	if info.Verification >= 3 {
		return false, ErrCodeVerifyTooManyTimes
	}

	// 比较验证码
	if inputCode != info.Code {
		// 验证失败，增加验证次数计数器
		info.Verification++
		c.store.Store(cacheKey, info)
		return false, ErrCodeIncorrect
	}

	// 验证成功后删除缓存
	c.store.Delete(cacheKey)
	return false, nil
}

func (c *LocalCodeCacheV2) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
