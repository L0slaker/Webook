package startup

import (
	"Prove/webook/internal/events"
	"Prove/webook/internal/events/article"
	"github.com/IBM/sarama"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9092"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

func NewConsumers(c *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c}
}
