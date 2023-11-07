package article

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, event ReadEvent) error
	ProduceReadEventV1(ctx context.Context, event ReadEventV1)
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type ReadEventV1 struct {
	Uids []int64
	Aids []int64
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(producer sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: producer,
	}
}

func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, event ReadEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (k *KafkaProducer) ProduceReadEventV1(ctx context.Context, event ReadEventV1) {
	//TODO implement me
	panic("implement me")
}
