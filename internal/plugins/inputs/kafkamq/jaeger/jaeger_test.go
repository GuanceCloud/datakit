// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jaeger handle Jaeger Spans from kafka.
package jaeger

import (
	"sync"
	"testing"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
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

var testSpanStr = `
{
	"traceId": "q9f9RaeqewMpQKWjWzj1Eg==",
	"spanId": "1f2n/O2U20w=",
	"operationName": "BaseController.readiness",
	"references": [{
		"traceId": "q9f9RaeqewMpQKWjWzj1Eg==",
		"spanId": "oD09Zb3lN1o="
	}],
	"startTime": "2023-12-01T10:47:45.853375888Z",
	"duration": "0.001245672s",
	"tags": [{
		"key": "thread.id",
		"vType": "INT64",
		"vInt64": "88"
	}, {
		"key": "thread.name",
		"vStr": "http-nio-9001-exec-8"
	}, {
		"key": "otel.library.name",
		"vStr": "io.opentelemetry.spring-webmvc-3.1"
	}, {
		"key": "otel.library.version",
		"vStr": "1.10.0"
	}, {
		"key": "internal.span.format",
		"vStr": "proto"
	}],
	"process": {
		"serviceName": "SUBSYS_ADGW_COREGW",
		"tags": [{
			"key": "clustername",
			"vStr": "zsc-k8sca"
		}, {
			"key": "hostname",
			"vStr": "adgw-coregw-py-6cbfcc9678-g2pmb"
		}, {
			"key": "ip",
			"vStr": "149.162.10.0"
		}, {
			"key": "jaeger.version",
			"vStr": "opentelemetry-java"
		}, {
			"key": "profile.env",
			"vStr": "zsc"
		}, {
			"key": "service.name",
			"vStr": "SUBSYS_ADGW_COREGW"
		}, {
			"key": "servicegroup",
			"vStr": "py"
		}, {
			"key": "telemetry.sdk.language",
			"vStr": "java"
		}, {
			"key": "telemetry.sdk.name",
			"vStr": "opentelemetry"
		}, {
			"key": "telemetry.sdk.version",
			"vStr": "1.10.0"
		}]
	}
}`

func BenchmarkParseToJaeger(b *testing.B) {
	feeder := dkio.NewMockedFeeder()
	mq := &Consumer{
		DKEndpoint: "http://localhost:9529/jaeger",
		Topics:     []string{"zsa"},
		Source:     "daoke",
		Thread:     32,
		dkURL:      "",
		feeder:     feeder,
		ptsLock:    sync.Mutex{},
		transport:  nil,
		wp:         nil,
	}
	if mq.Init() != nil {
		return
	}
	out := false
	go func() {
		for {
			if out {
				return
			}
			_, _ = feeder.NPoints(500, time.Second/10)
		}
	}()
	for i := 0; i < b.N; i++ {
		_ = mq.Process(&sarama.ConsumerMessage{Value: []byte(testSpanStr)})
	}
	out = true
}
