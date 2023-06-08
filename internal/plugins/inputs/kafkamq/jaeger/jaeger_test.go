// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jaeger handle Jaeger Spans from kafka.
package jaeger

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func TestProducerAndConsumer(t *testing.T) {
	// 创建Mock生产者
	producer := mocks.NewSyncProducer(t, nil)
	defer producer.Close()

	// 创建生产者消息
	message := &sarama.ProducerMessage{
		Topic: "test-topic",
		Value: sarama.StringEncoder("test-message"),
	}

	// 发送生产者消息
	producer.ExpectSendMessageAndSucceed()
	p, offset, err := producer.SendMessage(message)
	assert.NoError(t, err)
	t.Log(p, offset)
	// 创建Mock消费者
	consumer := mocks.NewConsumer(t, nil)
	defer consumer.Close()

	// 订阅主题
	pc := consumer.ExpectConsumePartition("test-topic", p, sarama.OffsetOldest)
	consumer.ConsumePartition("test-topic", p, sarama.OffsetOldest)
	pc.YieldMessage(&sarama.ConsumerMessage{Topic: "test-topic", Value: []byte("test-message"), Partition: 0})
	pc.YieldMessage(&sarama.ConsumerMessage{Topic: "test-topic", Value: []byte("test-message"), Partition: 0})
	// 接收消费者消息
	messages := pc.Messages()
	select {
	case msg := <-messages:
		assert.Equal(t, "test-message", string(msg.Value))
	default:
		t.Log("wait")
	}
	t.Log("ok")
}
