package cache

import (
	"Prove/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

type UserCache interface {
	Get(ctx context.Context, id int64) (*domain.User, error)
	Set(ctx context.Context, u *domain.User) error
}

// NewUserCache
// A 用到了 B，B 一定是接口 => 保证面向接口
// A 用到了 B，B 一定是 A 的字段 => 规避包变量、包方法，都非常缺乏拓展性
// A 用到了 B，A 绝对不初始化 B，而是外面注入 => 保持依赖注入和依赖反转
func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client: client,
		// 让别人传，或者自己设置
		//expiration: expiration,
		expiration: time.Minute * 15,
	}
}

// Get 只要error为nil，就认定缓存内有数据；如果没有数据，就返回一个特定的error
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	key := cache.key(id)
	value, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return &domain.User{}, err
	}
	var u *domain.User
	err = json.Unmarshal(value, u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u *domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}