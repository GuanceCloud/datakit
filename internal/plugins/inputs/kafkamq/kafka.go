// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  group
package kafkamq

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
	"golang.org/x/time/rate"
)

var (
	limiter *rate.Limiter // 令牌桶算法限速
	sample  *sampler      // 采样
)

// TopicProcess :process topic.
type TopicProcess interface {
	Init() error
	GetTopics() []string
	Process(msg *sarama.ConsumerMessage) error
}

type sampler struct {
	rate float64
}

func (s *sampler) sample() bool {
	num := rand.Intn(10) //nolint
	return num < int(s.rate*10)
}

type kafkaConsumer struct {
	process map[string]TopicProcess
	topics  []string
	addrs   []string
	groupID string
	stop    chan struct{}
	config  *sarama.Config
	ready   chan bool
}

func (kc *kafkaConsumer) limitAndSample(sec int, samplingRate float64) {
	if sec > 0 {
		// limit:速率  b:桶大小.
		limiter = rate.NewLimiter(rate.Limit(sec), sec*2)
	}
	if samplingRate < 1 && samplingRate > 0 {
		rand.Seed(time.Now().UnixNano())
		sample = &sampler{rate: samplingRate}
	}
}

func (kc *kafkaConsumer) registerP(handler TopicProcess) {
	if err := handler.Init(); err != nil {
		log.Errorf("init handler process err =%v", err)
		return
	}
	if handler != nil && len(handler.GetTopics()) > 0 {
		for _, t := range handler.GetTopics() {
			kc.process[t] = handler
			kc.topics = append(kc.topics, t)
		}
	}
	log.Infof("register topics=%v", handler.GetTopics())
}

func (kc *kafkaConsumer) start() {
	if len(kc.topics) == 0 {
		log.Warnf("topics len is 0")
		return
	}

	var group sarama.ConsumerGroup
	var err error
	var count int
	for {
		if count == 10 {
			log.Errorf("can not connect to kafka, consrmer exit")
			return
		}

		group, err = sarama.NewConsumerGroup(kc.addrs, kc.groupID, kc.config)
		if errors.Is(err, sarama.ErrOutOfBrokers) {
			group, err = UseSupportedVersions(kc.addrs, kc.groupID, kc.config)
			if group != nil {
				break
			}
		}
		if err != nil {
			log.Errorf("new group is err,restart count=%d ,addrs=[%v] err=%v", count, kc.addrs, err)
			time.Sleep(time.Second * 5)
			count++
			continue
		}
		break
	}
	if group == nil {
		log.Errorf("can not conn to kafka")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			if err := group.Consume(ctx, kc.topics, kc); err != nil {
				log.Errorf("error from consumer group =%v", err)
			}
			if ctx.Err() != nil {
				return
			}
			log.Infof("group:[%s] Re-election, consumer starts consuming messages", kc.groupID)
			time.Sleep(time.Second) // 防止频率太快 造成的日志太大.
			kafkaGroupElection.WithLabelValues().Add(1)
			kc.ready = make(chan bool)
		}
	}()

	<-kc.ready
	log.Infof("kafka: consumer group start...")
	select {
	case <-ctx.Done():
		log.Infof("terminating: context canceled")
	case <-kc.stop:
		log.Infof("kafka: exit")
	}
	_ = group.Close()
	cancel()
}

// Print : implement sarama STDLogger.
func (kc *kafkaConsumer) Print(v ...interface{}) {
	log.Debug(v)
}

// Printf : implement sarama STDLogger.
func (kc *kafkaConsumer) Printf(format string, v ...interface{}) {
	log.Debugf(format, v)
}

// Println : implement sarama STDLogger.
func (kc *kafkaConsumer) Println(v ...interface{}) {
	log.Debug(v)
}

// UseSupportedVersions :用户不提供版本信息，暴力破解版本.
func UseSupportedVersions(addrs []string, groupID string, config *sarama.Config) (sarama.ConsumerGroup, error) {
	var err error
	var group sarama.ConsumerGroup
	for i := len(sarama.SupportedVersions) - 1; i >= 0; i-- {
		config.Version = sarama.SupportedVersions[i]
		group, err = sarama.NewConsumerGroup(addrs, groupID, config)
		if err != nil {
			log.Errorf("new group is err,restart count=%d ,addrs=[%v] err=%v", i, addrs, err)
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}
	return group, err
}

// Setup : implement sarama ConsumerGroupHandler.
func (kc *kafkaConsumer) Setup(session sarama.ConsumerGroupSession) error {
	close(kc.ready)
	return nil
}

// Cleanup : implement sarama ConsumerGroupHandler.
func (kc *kafkaConsumer) Cleanup(session sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim : implement sarama ConsumerGroupHandler.
func (kc *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Infof("session was close")
				return nil
			}
			session.MarkMessage(msg, "")
			if msg == nil {
				log.Infof("message is nil,retrun")
				return nil
			}
			log.Debugf("message: %s", string(msg.Value))
			if sample != nil {
				if !sample.sample() {
					log.Debugf("sampler drop message")
					break
				}
			}
			topic := msg.Topic
			partition := fmt.Sprint(msg.Partition)
			startTime := time.Now()
			if p, ok := kc.process[topic]; ok {
				if err := p.Process(msg); err == nil {
					kafkaConsumeMessages.WithLabelValues(topic, partition, "ok").Add(1)
					processMessageCostVec.WithLabelValues(topic).Observe(float64(time.Since(startTime).Nanoseconds()))
				} else {
					kafkaConsumeMessages.WithLabelValues(topic, partition, "fail").Add(1)
				}
			} else {
				log.Warnf("can not find process for Topic:[%s]", topic)
			}
			if limiter != nil {
				_ = limiter.Wait(ctx)
			}
		case <-session.Context().Done():
			log.Infof("session context is close")
			return nil
		case <-kc.stop:
			return fmt.Errorf("datakit exit")
		}
	}
}
