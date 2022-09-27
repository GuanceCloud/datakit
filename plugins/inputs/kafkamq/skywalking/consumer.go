// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"context"
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
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

var (
	plugins        []string
	afterGatherRun itrace.AfterGatherHandler
	customerKeys   []string
	tags           map[string]string
	storage        *itrace.Storage
)

func InitOptions(pls []string, iStorage *itrace.Storage, closeResource map[string][]string,
	keepRareResource bool, sampler *itrace.Sampler, customerTags []string, itags map[string]string,
) {
	plugins = pls
	if iStorage != nil {
		if cache, err := cache.Open(iStorage.Path, &cache.Option{Capacity: int64(iStorage.Capacity) << 20}); err != nil {
			log.Errorf("### open cache %s with cap %dMB failed, cache.Open: %s", iStorage.Path, iStorage.Capacity, err)
		} else {
			iStorage.SetCache(cache)
			iStorage.RunStorageConsumer(log, parseSegmentObjectWrapper)
			storage = iStorage
			log.Infof("### open cache %s with cap %dMB OK", iStorage.Path, iStorage.Capacity)
		}
	}

	var afterGather *itrace.AfterGather
	if storage == nil {
		afterGather = itrace.NewAfterGather()
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithRetry(100 * time.Millisecond))
	}
	afterGatherRun = afterGather

	if len(closeResource) != 0 {
		iCloseResource := &itrace.CloseResource{}
		iCloseResource.UpdateIgnResList(closeResource)
		afterGather.AppendFilter(iCloseResource.Close)
	}

	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add rare resource keeper
	if keepRareResource {
		krs := &itrace.KeepRareResource{}
		krs.UpdateStatus(keepRareResource, time.Hour)
		afterGather.AppendFilter(krs.Keep)
	}

	// add sampler
	var isampler *itrace.Sampler
	if sampler != nil && (sampler.SamplingRateGlobal >= 0 && sampler.SamplingRateGlobal <= 1) {
		isampler = sampler
	} else {
		isampler = &itrace.Sampler{SamplingRateGlobal: 1}
	}
	afterGather.AppendFilter(isampler.Sample)

	customerKeys = customerTags
	tags = itags
}

type SkyConsumer struct {
	Topics    []string `toml:"topics"`
	Namespace string   `toml:"namespace"`

	handler *consumerGroupHandler
}

type consumerGroupHandler struct {
	topics map[string]string
	stop   chan struct{}
}

func (c *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			log.Debugf("Message claimed: timestamp = %v, topic = %s", msg.Timestamp, msg.Topic)
			session.MarkMessage(msg, "")
			if t, ok := c.topics[msg.Topic]; ok {
				switch t {
				case metric:
					metrics := &agentv3.JVMMetricCollection{}
					err := proto.Unmarshal(msg.Value, metrics)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}
					log.Debugf("unmarshal metrics is %+v", metrics)
					processMetrics(metrics)
				case profilings:
					profile := &profileV3.ThreadSnapshot{}
					err := proto.Unmarshal(msg.Value, profile)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}
					processProfile(profile)
				case segments:
					segment := &agentv3.SegmentObject{}
					err := proto.Unmarshal(msg.Value, segment)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}

					if storage == nil {
						parseSegmentObject(segment)
					} else {
						buf, err := proto.Marshal(segment)
						if err != nil {
							log.Error(err.Error())
							return err
						}
						param := &itrace.TraceParameters{Meta: &itrace.TraceMeta{Buf: buf}}
						if err = storage.Send(param); err != nil {
							log.Error(err.Error())

							return err
						}
					}
				case managements:
					instance := &managementV3.InstanceProperties{}
					err := proto.Unmarshal(msg.Value, instance)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}
					log.Debugf("unmarshal instance is= %+v", instance)
				case meters:
					meters := &agentv3.MeterData{}
					err := proto.Unmarshal(msg.Value, meters)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}
					log.Debugf("unmarshal Metrer is= %+v", meters)
				case logging:
					log.Debugf("logging")
					pLog := &loggingv3.LogData{}
					err := proto.Unmarshal(msg.Value, pLog)
					if err != nil {
						log.Errorf("unmarshal err=%v", err)
						break
					}
					processLog(pLog)
				}
			}
		case <-c.stop:
			return nil
		}
	}
}

func (mq *SkyConsumer) SaramaConsumerGroup(addr string, groupID string) {
	log = logger.SLogger("kafkamq")
	config := sarama.NewConfig() // auto commit
	config.Consumer.Return.Errors = false
	config.Version = sarama.V0_10_2_0                     // specify appropriate version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 未找到组消费位移的时候从哪边开始消费
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
	log.Infof("skywalking consumer addr= %s  groupID=%s", addr, groupID)
	var group sarama.ConsumerGroup
	var err error
	var count int
	for {
		if count == 10 {
			log.Errorf("can not connect to kafka, consrmer exit")
			return
		}
		group, err = sarama.NewConsumerGroup([]string{addr}, groupID, config)
		if err != nil {
			log.Errorf("new group is err,restart count=%d , err=%v", count, err)
			time.Sleep(time.Second * 10)
			count++
			continue
		}
		break
	}

	defer func() { _ = group.Close() }()

	// Track errors
	go func() {
		for err := range group.Errors() {
			log.Errorf("group has err=%v", err)
		}
	}()
	log.Infof("Consumed start")
	// Iterate over consumer sessions.
	ctx := context.Background()
	for {
		handler := &consumerGroupHandler{topics: formatTopics, stop: make(chan struct{}, 1)}
		mq.handler = handler
		log.Infof("start consumer....")
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims.
		err := group.Consume(ctx, topics, handler)
		if err != nil {
			log.Errorf("group begin consume is err=%v", err)
		}
		time.Sleep(time.Second * 5)
	}
}

func (mq *SkyConsumer) Stop() {
	if mq.handler != nil {
		mq.handler.stop <- struct{}{}
	}
}
