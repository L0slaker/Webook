package saramax

import (
	"Prove/webook/pkg/logger"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"time"
)

// BatchHandler 批量消费
type BatchHandler[T any] struct {
	l  logger.LoggerV1
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 最大 goroutine 数量，可以考虑使用 option 模式来设置
	batchSize int
	// 超时时间，可以考虑使用 option 模式来设置
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:             l,
		fn:            fn,
		batchSize:     10,
		batchDuration: time.Second,
	}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, b.batchSize)
		ts := make([]T, 0, b.batchSize)
		for i := 0; i < b.batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					// 消费者已被关闭
					cancel()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, t)
				if err != nil {
					b.l.Error("反序列化失败！",
						logger.Error(err),
						logger.Int64("offset", msg.Offset),
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		// 如果一条消息都没拿到，直接跳过
		if len(msgs) == 0 {
			continue
		}
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error("调用业务批量接口失败！", logger.Error(err))
			// 记录整个批次
			// 继续往前消费
		}
		for _, msg := range msgs {
			session.MarkMessage(msg, "")
		}
	}
}
