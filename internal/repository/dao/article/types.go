package article

import (
	"context"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	GetByAuthorId(ctx context.Context, authorId int64, offset, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, art Article) error
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error)
}
