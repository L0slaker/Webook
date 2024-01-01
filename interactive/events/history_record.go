package events

import (
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	//repo   repository2.HistoryRecordRepository
}

func NewHistoryReadEventConsumer(client sarama.Client, l logger.LoggerV1) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		l:      l,
	}
}

func (h *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("history_record", h.client)
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
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//return h.repo.AddRecord(ctx, t.Uid, t.Aid)
	panic("implement me")
}
