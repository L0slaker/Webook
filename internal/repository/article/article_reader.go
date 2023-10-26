package article

import (
	"Prove/webook/internal/domain"
	"context"
)

type ArticleReaderRepository interface {
	//Save 相当于 Upsert 的语义
	Save(ctx context.Context, article domain.Article) (int64, error)
}
