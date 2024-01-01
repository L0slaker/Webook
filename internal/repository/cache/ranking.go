package cache

import (
	"Prove/webook/internal/domain"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client redis.Cmdable
	key    string
}

func NewRankingRedisCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{
		client: client,
		key:    "ranking",
	}
}

func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	// 不缓存内容
	for i := 0; i < len(arts); i++ {
		arts[i].Content = ""
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 将过期时间设置为永不过期，避免数据库不可用的情况下，无法获取排行榜
	return r.client.Set(ctx, r.key, val, 0).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}
