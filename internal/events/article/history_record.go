package article

import (
	"Prove/webook/internal/repository"
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	repo   repository.InteractiveRepository
}

func NewHistoryReadEventConsumer(client sarama.Client, l logger.LoggerV1,
	repo repository.InteractiveRepository) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (h *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", h.client)
	if err != nil {
		return err
	}
	go func() {
		e := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](h.l, h.Consume))
		if e != nil {
			h.l.Error("退出了消费循环异常", logger.Error(e))
		}
	}()
	return err
}

func (h *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return h.repo.AddRecord(ctx, t.Uid, t.Aid)
}
