package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var addrs = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	//// 随机
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	//// 一致性哈希，CRC16 算法
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	//// 根据 message 的 partition 字段来选择
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner
	//// 根据 key 的哈希值来筛选一个，默认值
	cfg.Producer.Partitioner = sarama.NewHashPartitioner

	producer, err := sarama.NewSyncProducer(addrs, cfg)
	assert.NoError(t, err)
	for i := 0; i < 1000; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "read_article",
			// 消息数据本体
			Value: sarama.StringEncoder(`{"aid":2,"uid":123}`),
			// 在生产者和消费者之间传递
			//Headers: []sarama.RecordHeader{
			//	{
			//		Key:   []byte("trace_id"),
			//		Value: []byte("123456"),
			//	},
			//},
			// 只作用于发送过程
			//Metadata: "this is metadata",
		})
		assert.NoError(t, err)
	}
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 发送成功的部分
	cfg.Producer.Return.Successes = true
	// 发送不成功的部分
	cfg.Producer.Return.Errors = true
	// 指定 acks：（从上到下，性能变差，但数据可靠性上升）
	// 0：客户端发一次，不需要服务端的确认
	// 1：客户端发送，并且需要服务端写入到主分区
	// -1：客户端发送，并且需要服务端同步到所有的 ISR 上
	cfg.Producer.RequiredAcks = 0

	producer, err := sarama.NewAsyncProducer(addrs, cfg)
	require.NoError(t, err)
	msgCh := producer.Input()
	msgCh <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("德玛西亚！"),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("trace_id"),
				Value: []byte("123456789"),
			},
		},
		Metadata:  "this is metadata async_send",
		Partition: 0,
	}

	errCh := producer.Errors()
	successCh := producer.Successes()

	for {
		select {
		case e := <-errCh:
			t.Log("发送出现了问题！", e.Err, e.Msg.Value)
		case <-successCh:
			t.Log("发送成功！")
		}
		time.Sleep(time.Second * 30)
	}
}
