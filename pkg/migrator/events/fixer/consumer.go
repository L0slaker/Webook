package fixer

import (
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/migrator"
	"Prove/webook/pkg/migrator/events"
	"Prove/webook/pkg/migrator/fixer"
	"Prove/webook/pkg/saramax"
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](client sarama.Client, l logger.LoggerV1,
	src, dst *gorm.DB, topic string) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{c.topic},
			saramax.NewHandler[events.InconsistentEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *Consumer[T]) Consume(msg *sarama.ConsumerMessage, evt events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch evt.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, evt.ID)
	case "DST":
		return c.dstFirst.Fix(ctx, evt.ID)
	}
	return errors.New("未知的校验方向")
}
