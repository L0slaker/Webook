package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LocalCodeCache struct {
	store *sync.Map // 存储验证码的本地缓存
}

type CodeInfo struct {
	Code         string
	Expiration   time.Time
	Verification int
}

func NewLocalCodeCache(store *sync.Map) CodeCache {
	return &LocalCodeCache{
		store: store,
	}
}

func (c *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	cacheKey := c.key(biz, phone)

	val, ok := c.store.Load(cacheKey)
	if ok {
		info := val.(CodeInfo)
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
	info := CodeInfo{
		Code:         code,
		Expiration:   time.Now(),
		Verification: 0,
	}
	c.store.Store(cacheKey, info)
	return nil
}

func (c *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	cacheKey := c.key(biz, phone)

	// 获取存储的验证码
	val, ok := c.store.Load(cacheKey)
	if !ok {
		return false, nil
	}

	info := val.(CodeInfo)
	// 检查验证次数
	if info.Verification >= 3 {
		return false, ErrCodeVerifyTooManyTimes
	}

	// 比较验证码
	if inputCode != info.Code {
		// 验证失败，增加验证次数计数器
		info.Verification++
		c.store.Store(cacheKey, info)
		return true, ErrCodeIncorrect
	}

	// 验证成功后删除缓存
	c.store.Delete(cacheKey)
	return false, nil
}

func (c *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
