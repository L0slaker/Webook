package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	p     sarama.SyncProducer
	topic string
}

func NewSaramaProducer(p sarama.SyncProducer, topic string) *SaramaProducer {
	return &SaramaProducer{
		p:     p,
		topic: topic,
	}
}

// ProduceInconsistentEvent 当需要校验的时候，就发送一条消息
func (s *SaramaProducer) ProduceInconsistentEvent(ctx context.Context, event InconsistentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = s.p.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(data),
	})
	return err
}
