// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package custom kafka MQ.
package custom

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/Shopify/sarama"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

var (
	log     = logger.DefaultSLogger("kafkamq_custom")
	MsgType = "kafka"
)

// Custom : 自定义消息的处理对象.
type Custom struct {
	GroupID      string   `toml:"group_id"`
	LogTopics    []string `toml:"log_topics"`
	LogPl        string   `toml:"log_pl"`
	MetricTopics []string `toml:"metric_topic"`
	MetricPl     string   `toml:"metric_pl"`
	SpiltBody    bool     `toml:"spilt_json_body"`

	logTopicsMap    map[string]string
	metricTopicsMap map[string]string
	feeder          dkio.Feeder
}

// Init :初始化消息.
func (mq *Custom) Init() error {
	log = logger.SLogger("kafkamq.custom")

	mq.withTopic(mq.LogTopics, mq.LogPl, mq.MetricTopics, mq.MetricPl)
	if len(mq.LogTopics) == 0 && len(mq.MetricTopics) == 0 {
		log.Warnf("no custom topics")
		return fmt.Errorf("no custom topics")
	}
	return nil
}

func (mq *Custom) SetFeeder(feeder dkio.Feeder) {
	mq.feeder = feeder
}

// GetTopics :TopicProcess implement.
func (mq *Custom) GetTopics() []string {
	return append(mq.LogTopics, mq.MetricTopics...)
}

// Process :TopicProcess implement.
func (mq *Custom) Process(msg *sarama.ConsumerMessage) {
	if mq.feeder == nil {
		log.Warnf("feeder is nil,return")
		return
	}
	pl := ""
	msgs := make([]string, 0)
	category := point.Logging
	opts := make([]point.Option, 0)
	if mq.SpiltBody {
		is := make([]interface{}, 0)
		err := json.Unmarshal(msg.Value, &is)
		if err != nil {
			log.Warnf("Unmarshal err=%v", err)
			return
		}
		for _, i := range is {
			bts, err := json.Marshal(i)
			if err != nil {
				log.Warnf("marshal err=%v", err)
				return
			}
			m := strings.ReplaceAll(string(bts), "\n", " ")
			log.Debugf("kafka_message is %s", m)
			msgs = append(msgs, m)
		}
	} else {
		newMessage := strings.ReplaceAll(string(msg.Value), "\n", " ")
		log.Debugf("kafka_message is:  %s", newMessage)
		msgs = append(msgs, newMessage)
	}

	for _, msgStr := range msgs {
		tags := map[string]string{"type": MsgType}
		msgLen := len(msgStr)
		fields := map[string]interface{}{pipeline.FieldStatus: pipeline.DefaultStatus, "message_len": msgLen}
		if p, ok := mq.logTopicsMap[msg.Topic]; ok {
			pl = p
			fields[pipeline.FieldMessage] = msgStr
		}
		if p, ok := mq.metricTopicsMap[msg.Topic]; ok {
			pl = p
			category = point.Metric
			tags[pipeline.FieldMessage] = msgStr
		}
		if pl == "" {
			log.Warnf("can not find [%s] pipeline script", msg.Topic)
			return
		}

		pt := point.NewPointV2([]byte(msg.Topic),
			append(point.NewTags(tags), point.NewKVs(fields)...),
			opts...)

		err := mq.feeder.Feed(msg.Topic, category, []*point.Point{pt}, &dkio.Option{PlScript: map[string]string{msg.Topic: pl}})
		if err != nil {
			log.Warnf("feed io err=%v", err)
		}
	}
}

func (mq *Custom) withTopic(logTopics []string, lpl string, metricTopics []string, mpl string) {
	if mq.logTopicsMap == nil {
		mq.logTopicsMap = make(map[string]string)
	}
	if mq.metricTopicsMap == nil {
		mq.metricTopicsMap = make(map[string]string)
	}
	for _, topic := range logTopics {
		mq.logTopicsMap[topic] = lpl
	}
	for _, topic := range metricTopics {
		mq.metricTopicsMap[topic] = mpl
	}
}
