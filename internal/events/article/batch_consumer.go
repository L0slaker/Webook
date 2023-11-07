package article

import (
	"Prove/webook/internal/repository"
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	repo   repository.InteractiveRepository
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, l logger.LoggerV1, repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (i *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		e := cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](i.l, i.Consume))
		if e != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (i *InteractiveReadEventBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := i.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		i.l.Error("批量增加阅读计数失败！",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil
}
