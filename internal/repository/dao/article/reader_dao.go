package article

import (
	"context"
	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	// UpsertV1 模拟完全不同的表
	UpsertV1(ctx context.Context, art PublishedArticle) error
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	panic("implement me")
}
