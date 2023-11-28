package cache

import (
	"context"
	"github.com/ecodeclub/ekit"
	"time"
)

type Cache interface {
	Get(ctx context.Context) ekit.AnyValue
	Set(ctx context.Context, key string, val any, exp time.Duration) error
}
