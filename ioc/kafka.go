package ioc

import (
	"Prove/webook/internal/events"
	"Prove/webook/internal/events/article"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	// 配置了生产者（Producer）返回成功消息
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	// 使用 viper 库来从配置文件中解析 "kafka" 部分的配置信息，并将其映射到 cfg 变量中
	if err := viper.UnmarshalKey("kafka", &cfg); err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
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