package cache

import (
	"Prove/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, authorId int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, authorId int64) error
	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error) {
	bs, err := r.client.Get(ctx, r.firstPageKey(authorId)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bs, &arts)
	return arts, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, authorId int64, arts []domain.Article) error {
	// 只需要缓存摘要，而不是缓存内容
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.firstPageKey(authorId), data, time.Minute*10).Err()
}

func (r *RedisArticleCache) DelFirstPage(ctx context.Context, authorId int64) error {
	return r.client.Del(ctx, r.firstPageKey(authorId)).Err()
}

func (r *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.authorArtKey(art.Id), data, time.Minute).Err()
}

func (r *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	data, err := r.client.Get(ctx, r.authorArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (r *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 设置过期时间
	return r.client.Set(ctx, r.readerArtKey(art.Id), data, time.Minute*30).Err()
}

func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	data, err := r.client.Get(ctx, r.readerArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (r *RedisArticleCache) firstPageKey(authorId int64) string {
	return fmt.Sprintf("article:first_page:%d", authorId)
}

// 作者端的缓存设置
func (r *RedisArticleCache) authorArtKey(id int64) string {
	return fmt.Sprintf("article:author:%d", id)
}

// 读者端的缓存设置
func (r *RedisArticleCache) readerArtKey(id int64) string {
	return fmt.Sprintf("article:reader:%d", id)
}
