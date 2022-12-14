// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"context"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/skywalkingapi"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/agent/v3"
	profileV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/profile/v3"
	loggingv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/logging/v3"
	managementV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/management/v3"

	"github.com/Shopify/sarama"
	"google.golang.org/protobuf/proto"
)

var (
	log         = logger.DefaultSLogger("kafkamq")
	metric      = "skywalking-metrics"
	profilings  = "skywalking-profilings"
	segments    = "skywalking-segments"
	managements = "skywalking-managements"
	meters      = "skywalking-meters"
	logging     = "skywalking-logging"
)

var api *skywalkingapi.SkyAPI

type SkyConsumer struct {
	Topics    []string `toml:"topics"`
	Namespace string   `toml:"namespace"`

	stop chan struct{}
}

type consumerGroupHandler struct {
	topics map[string]string
	ready  chan bool
}

func (c *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

func (c *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if api == nil {
		log.Errorf("skywalking api is nil")
		return nil
	}
	for {
		select {
		case msg := <-claim.Messages():
			if msg == nil {
				log.Infof("session was close")
				return nil
			}
			log.Debugf("Message claimed: timestamp = %v, topic = %s", msg.Timestamp, msg.Topic)
			session.MarkMessage(msg, "")
			if t, ok := c.topics[msg.Topic]; ok {
				switch t {
				case metric:
					metrics := &agentv3.JVMMetricCollection{}
					err := proto.Unmarshal(msg.Value, metrics)
					if err != nil {
						log.Warnf("unmarshal err=%v", err)
						break
					}
					log.Debugf("unmarshal metrics is %+v", metrics)
					api.ProcessMetrics(metrics)
				case profilings:
					profile := &profileV3.ThreadSnapshot{}
					err := proto.Unmarshal(msg.Value, profile)
					if err != nil {
						log.Warnf("unmarshal err=%v", err)
						break
					}
					api.ProcessProfile(profile)
				case segments:
					segment := &agentv3.SegmentObject{}
					err := proto.Unmarshal(msg.Value, segment)
					if err != nil {
						log.Warnf("Marshal err =%v , and val is [%s]", err, string(msg.Value))
						break
					}
					api.ProcessSegment(segment)
				case managements:
					instance := &managementV3.InstanceProperties{}
					err := proto.Unmarshal(msg.Value, instance)
					if err != nil {
						log.Warnf("unmarshal err=%v", err)
						break
					}
					log.Debugf("unmarshal instance is= %+v", instance)
				case meters:
					meters := &agentv3.MeterData{}
					err := proto.Unmarshal(msg.Value, meters)
					if err != nil {
						log.Warnf("Marshal err =%v , and val is [%s]", err, string(msg.Value))
						break
					}
					log.Debugf("unmarshal Metrer is= %+v", meters)
				case logging:
					pLog := &loggingv3.LogData{}
					err := proto.Unmarshal(msg.Value, pLog)
					if err != nil {
						log.Warnf("unmarshal err=%v", err)
						break
					}
					api.ProcessLog(pLog)
				}
			}
		case <-session.Context().Done():
			log.Infof("session context is close")
			return nil
		}
	}
}

//nolint:lll
func (mq *SkyConsumer) SaramaConsumerGroup(addrs []string, groupID string, skyapi *skywalkingapi.SkyAPI, version sarama.KafkaVersion, balance sarama.BalanceStrategy) {
	if skyapi != nil {
		api = skyapi
	} else {
		log.Errorf("skywalking api is nil")
		return
	}
	log = logger.SLogger("kafkamq")
	mq.stop = make(chan struct{}, 1)
	config := sarama.NewConfig() // auto commit
	config.Consumer.Return.Errors = false
	config.Version = version                              // specify appropriate version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 未找到组消费位移的时候从哪边开始消费
	config.Consumer.Group.Rebalance.Strategy = balance
	formatTopics := make(map[string]string)
	topics := make([]string, 0)
	for _, topic := range mq.Topics {
		if mq.Namespace != "" {
			formatTopics[mq.Namespace+"-"+topic] = topic
		} else {
			formatTopics[topic] = topic
		}
		topics = append(topics, topic)
	}
	log.Debugf("topic is %v", topics)
	log.Infof("skywalking consumer addr= %s  groupID=%s", addrs, groupID)
	var group sarama.ConsumerGroup
	var err error
	var count int
	for {
		if count == 10 {
			log.Errorf("can not connect to kafka, consrmer exit")
			return
		}
		group, err = sarama.NewConsumerGroup(addrs, groupID, config)
		if err != nil {
			log.Errorf("new group is err,restart count=%d , err=%v", count, err)
			time.Sleep(time.Second * 10)
			count++
			continue
		}
		break
	}

	// Iterate over consumer sessions.
	ctx, cancel := context.WithCancel(context.Background())

	handler := &consumerGroupHandler{topics: formatTopics, ready: make(chan bool)}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims.
			if err := group.Consume(ctx, topics, handler); err != nil {
				log.Errorf("Error from consumer: %v", err)
			}
			// check if context was canceled, signaling that the consumer should stop.
			if ctx.Err() != nil {
				return
			}
			time.Sleep(time.Second)
			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready // Await till the consumer has been set up
	log.Infof("Sarama consumer up and running!...")

	select {
	case <-ctx.Done():
		log.Infof("terminating: context canceled")
	case <-mq.stop:
		log.Infof("consumer stop")
	}

	cancel()
	wg.Wait()
	if err = group.Close(); err != nil {
		log.Errorf("Error closing client: %v", err)
	}
}

func (mq *SkyConsumer) Stop() {
	mq.stop <- struct{}{}
	if api != nil {
		api.StopStorage()
	}
}
