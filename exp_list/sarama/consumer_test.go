package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

var groupId = "test_consumer"

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()

	consumer, err := sarama.NewConsumerGroup(addrs, groupId, cfg)
	require.NoError(t, err)
	err = consumer.Consume(context.Background(), []string{"test_topic"}, testConsumerGroupHandler{})
	t.Log(err)
}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// topic => 偏移量
	partitions := session.Claims()["test_topic"]
	for _, part := range partitions {
		session.ResetOffset("test_topic", part, sarama.OffsetOldest, "")
	}
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("clean_up")
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	// 这种做法有很大风险：如果生产的消费并发量很大
	// 而我们这里没有限制goroutine的数量，可能会导致系统安全问题
	//for msg := range msgs {
	//	m1 := msg
	//	go func() {
	//		// 消费msg
	//		log.Println(string(m1.Value))
	//		// 提交
	//		session.MarkMessage(m1, "")
	//	}()
	//}
	const maxSize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		done := false
		for i := 0; i < maxSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					cancel()
					// 消费者已被关闭
					return nil
				}
				last = msg
				eg.Go(func() error {
					// 消费
					time.Sleep(time.Second)
					// 重试
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			// 记录日志
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}

// 返回一个只读的channel
func ChannelV1() <-chan int {
	panic("implement me")
}

// 返回只写的channel
func ChannelV2() chan<- int {
	panic("implement me")
}

// 返回可读可写的channel
func ChannelV3() chan int {
	panic("implement me")
}

type MyBizMsg struct {
	Name string
}
