package cache

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	_ "github.com/hashicorp/golang-lru/v2/expirable"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache *lru.Cache
	mu    *sync.Mutex
}

type CodeInfo struct {
	Code         string
	Expiration   time.Time
	Verification int
}

func NewLocalCodeCache(size int, mu *sync.Mutex) CodeCache {
	//cache := expirable.NewLRU(size, nil, time.Minute*10)
	cache, _ := lru.New(size)
	return &LocalCodeCache{
		cache: cache,
		mu:    mu,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	cacheKey := l.key(biz, phone)

	val, ok := l.cache.Get(cacheKey)
	if ok {
		info := val.(CodeInfo)
		elapsed := time.Since(info.Expiration)
		if elapsed < 60*time.Second {
			return ErrCodeSendTooMany
		}
		if elapsed >= 600*time.Second {
			return ErrCodeSendExpired
		}
	}

	l.cache.Add(cacheKey, CodeInfo{
		Code:         code,
		Expiration:   time.Now(),
		Verification: 0,
	})

	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cacheKey := l.key(biz, phone)

	if cachedCode, exists := l.cache.Get(cacheKey); exists {
		info := cachedCode.(CodeInfo)
		// 检查验证次数
		if info.Verification >= 3 {
			return false, ErrCodeVerifyTooManyTimes
		}

		// 比较验证码
		if inputCode != info.Code {
			// 验证失败，增加验证次数计数器
			info.Verification++
			l.cache.Add(cacheKey, info)
			return false, ErrCodeIncorrect
		}

		l.cache.Remove(cacheKey)
		return true, nil
	}
	return false, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
