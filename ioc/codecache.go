package ioc

import (
	"Prove/webook/internal/repository/cache"
	"github.com/redis/go-redis/v9"
)

func InitCodeCache(client redis.Cmdable) cache.CodeCache {
	// 暂时使用 redis 的机制
	return cache.NewRedisCodeCache(client)
}
