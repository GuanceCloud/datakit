// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package custom kafka MQ.
package custom

import (
	"context"
	"errors"
	"math/rand" //nolint
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"golang.org/x/time/rate"
)

var (
	log     = logger.DefaultSLogger("kafkamq_custom")
	limiter *rate.Limiter
	sample  *sampler
)

type Custom struct {
	GroupID      string   `toml:"group_id"`
	LogTopics    []string `toml:"log_topics"`
	LogPl        string   `toml:"log_pl"`
	MetricTopics []string `toml:"metric_topic"`
	MetricPl     string   `toml:"metric_pl"`
	LimitSec     int      `toml:"limit_sec"` // 令牌桶速率限制，每秒多少个.
	SamplingRate float64  `toml:"sampling_rate"`
	//	SampleRote   float64  // 0~1.0
	stop chan struct{}
}

func (c *Custom) SaramaConsumerGroup(addrs []string, config *sarama.Config) {
	log = logger.SLogger("kafkamq_custom")
	sarama.Logger = c
	if c.LimitSec > 0 {
		// limit:速率  b:桶大小.
		limiter = rate.NewLimiter(rate.Limit(c.LimitSec), c.LimitSec*2)
	}
	if c.SamplingRate < 1 && c.SamplingRate > 0 {
		rand.Seed(time.Now().UnixNano())
		sample = &sampler{rate: c.SamplingRate}
	}

	c.stop = make(chan struct{}, 1)
	var group sarama.ConsumerGroup
	var err error
	var count int
	for {
		if count == 10 {
			log.Errorf("can not connect to kafka, consrmer exit")
			return
		}

		group, err = sarama.NewConsumerGroup(addrs, c.GroupID, config)
		if errors.Is(err, sarama.ErrOutOfBrokers) {
			group, err = UseSupportedVersions(addrs, c.GroupID, config)
			if group != nil {
				break
			}
		}
		if err != nil {
			log.Errorf("new group is err,restart count=%d ,addrs=[%v] err=%v", count, addrs, err)
			time.Sleep(time.Second * 10)
			count++
			continue
		}
		break
	}
	topics := make([]string, 0)

	topics = append(topics, c.LogTopics...)
	topics = append(topics, c.MetricTopics...)

	// Iterate over consumer sessions.
	ctx, cancel := context.WithCancel(context.Background())

	handler := &consumerGroupHandler{ready: make(chan bool)}
	handler.withTopic(c.LogTopics, c.LogPl, c.MetricTopics, c.MetricPl)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	log.Infof("custom is run with topics =[%+v]", topics)
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
			time.Sleep(time.Second) // 防止频率太快 造成的日志太大.
			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready // Await till the consumer has been set up
	log.Infof("Sarama consumer up and running!...")

	select {
	case <-ctx.Done():
		log.Infof("terminating: context canceled")
	case <-c.stop:
		log.Infof("consumer stop")
	}

	cancel()
	wg.Wait()
	if err = group.Close(); err != nil {
		log.Errorf("Error closing client: %v", err)
	}
}

func (c *Custom) Stop() {
	if c.stop != nil {
		c.stop <- struct{}{}
	}
}

func (c *Custom) Print(v ...interface{}) {
	log.Debug(v)
}

func (c *Custom) Printf(format string, v ...interface{}) {
	log.Debugf(format, v)
}

func (c *Custom) Println(v ...interface{}) {
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
			time.Sleep(time.Second * 10)
		} else {
			break
		}
	}
	return group, err
}

type sampler struct {
	rate float64
}

func (s *sampler) sample() bool {
	num := rand.Intn(10) //nolint
	return num < int(s.rate*10)
}

type consumerGroupHandler struct {
	// 暂时先支持 log 和 metric 两种数据.
	logTopics    map[string]string
	metricTopics map[string]string
	ready        chan bool
}

func (c *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

func (c *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	for {
		select {
		case msg := <-claim.Messages():
			log.Debugf("partition=%d", claim.Partition())
			if msg == nil {
				return nil
			}
			log.Debugf("message topic =%s", msg.Topic)
			session.MarkMessage(msg, "")

			if sample != nil {
				if !sample.sample() {
					log.Debugf("sampler drop message")
					break
				}
			}
			pl := ""
			category := ""
			if p, ok := c.logTopics[msg.Topic]; ok {
				pl = p
				category = datakit.Logging
			}
			if p, ok := c.metricTopics[msg.Topic]; ok {
				pl = p
				category = datakit.Metric
			}
			if pl == "" || category == "" {
				log.Warnf("can not find [%s] pipeline script", msg.Topic)
				break
			}
			// 把换行符替换成空格.
			newMessage := strings.ReplaceAll(string(msg.Value), "\n", " ")
			log.Debugf("kafka_message is:  %s", newMessage)
			msgLen := len(newMessage)

			pt, err := point.NewPoint(
				msg.Topic,
				map[string]string{"type": "kafka", pipeline.FieldMessage: newMessage},
				map[string]interface{}{pipeline.FieldStatus: pipeline.DefaultStatus, "message_len": msgLen},
				&point.PointOption{
					Time:     time.Now(),
					Category: category,
				})
			if err != nil {
				log.Warnf("make point err=%v", err)
				break
			}

			err = dkio.Feed(msg.Topic, category, []*point.Point{pt}, &dkio.Option{PlScript: map[string]string{msg.Topic: pl}})
			if err != nil {
				log.Warnf("feed io err=%v", err)
			}
			if limiter != nil {
				_ = limiter.Wait(ctx)
			}
		case <-session.Context().Done():
			log.Infof("session context is close")
			return nil
		}
	}
}

func (c *consumerGroupHandler) withTopic(logTopics []string, lpl string, metricTopics []string, mpl string) {
	if c.logTopics == nil {
		c.logTopics = make(map[string]string)
	}
	if c.metricTopics == nil {
		c.metricTopics = make(map[string]string)
	}
	for _, topic := range logTopics {
		c.logTopics[topic] = lpl
	}
	for _, topic := range metricTopics {
		c.metricTopics[topic] = mpl
	}
}
