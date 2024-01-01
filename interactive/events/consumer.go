package events

import (
	"Prove/webook/interactive/repository"
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	repo   repository.InteractiveRepository
}

func NewInteractiveReadEventConsumer(client sarama.Client, l logger.LoggerV1,
	repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		e := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if e != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 不是幂等的
func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", t.Aid)
}
