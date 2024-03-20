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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/logger"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/IBM/sarama"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/worker"
)

var (
	log     = logger.DefaultSLogger("kafkamq_custom")
	MsgType = "kafka"
)

// Custom 自定义消息的处理对象.
type Custom struct {
	// old config
	LogTopics    []string `toml:"log_topics"`
	LogPl        string   `toml:"log_pl"`
	MetricTopics []string `toml:"metric_topic"`
	MetricPl     string   `toml:"metric_pl"`

	// since v1.6.0
	LogTopicsMap    map[string]string `toml:"log_topic_map"`
	MetricTopicsMap map[string]string `toml:"metric_topic_map"`
	RumTopicsMap    map[string]string `toml:"rum_topic_map"`
	SpiltTopicsMap  map[string]bool   `toml:"spilt_topic_map"`
	SpiltBody       bool              `toml:"spilt_json_body"`
	Thread          int               `toml:"thread"`
	wp              *worker.WorkerPool
	feeder          dkio.Feeder
	Tagger          datakit.GlobalTagger
}

// Init 初始化消息.
func (mq *Custom) Init() error {
	log = logger.SLogger("kafkamq.custom")

	mq.initMap()

	if len(mq.LogTopicsMap) == 0 && len(mq.MetricTopicsMap) == 0 && len(mq.RumTopicsMap) == 0 {
		log.Warnf("no custom topics")
		return fmt.Errorf("no custom topics")
	}
	mq.wp = worker.NewWorkerPool(mq.DoMsg, mq.Thread)
	mq.Tagger = datakit.DefaultGlobalTagger()
	return nil
}

func (mq *Custom) SetFeeder(feeder dkio.Feeder) {
	mq.feeder = feeder
}

// GetTopics TopicProcess implement.
func (mq *Custom) GetTopics() []string {
	topics := make([]string, 0)
	if len(mq.LogTopicsMap) > 0 {
		for t := range mq.LogTopicsMap {
			topics = append(topics, t)
		}
	}
	if len(mq.MetricTopicsMap) > 0 {
		for t := range mq.MetricTopicsMap {
			topics = append(topics, t)
		}
	}
	if len(mq.RumTopicsMap) > 0 {
		for t := range mq.RumTopicsMap {
			topics = append(topics, t)
		}
	}

	return topics
}

// Process TopicProcess implement.
func (mq *Custom) Process(msg *sarama.ConsumerMessage) error {
	if mq.wp != nil {
		f := mq.wp.GetWorker()
		go func(message *sarama.ConsumerMessage) {
			err := f(message)
			if err != nil {
				log.Errorf("process message err=%v", err)
			}
			mq.wp.PutWorker(f)
		}(msg)
		return nil
	} else {
		return mq.DoMsg(msg)
	}
}

func (mq *Custom) DoMsg(msg *sarama.ConsumerMessage) error {
	var (
		err              error
		topic            = msg.Topic
		msgs             = make([]string, 0)
		category         = point.Logging
		opts             = make([]point.Option, 0)
		topicToSpilt, ok = mq.SpiltTopicsMap[topic]
	)
	if mq.feeder == nil {
		err = fmt.Errorf("feeder is nil,return")
		log.Warn(err)
		return err
	}
	if !ok {
		topicToSpilt = mq.SpiltBody
	}
	if topicToSpilt {
		is := make([]interface{}, 0)
		err := json.Unmarshal(msg.Value, &is)
		if err != nil {
			log.Warnf("Unmarshal err=%v", err)
			return err
		}
		for _, i := range is {
			bts, err := json.Marshal(i)
			if err != nil {
				log.Warnf("marshal err=%v", err)
				return err
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
		plMap := map[string]string{}
		if p, ok := mq.LogTopicsMap[topic]; ok {
			fields[pipeline.FieldMessage] = msgStr
			plMap[topic] = p
			opts = append(point.DefaultLoggingOptions(), point.WithExtraTags(mq.Tagger.HostTags()))
		}
		if p, ok := mq.MetricTopicsMap[topic]; ok {
			category = point.Metric
			tags[pipeline.FieldMessage] = msgStr
			plMap[topic] = p
			opts = append(point.DefaultMetricOptions(), point.WithExtraTags(mq.Tagger.HostTags()))
		}

		if p, ok := mq.RumTopicsMap[topic]; ok {
			category = point.RUM
			fields[pipeline.FieldMessage] = msgStr
			tags["app_id"] = topic     // 这里只是一个占位符，没有实际意义，但是 rum 数据中的 app_id 是必须要有的字段。
			plMap[topic+"_"+topic] = p // 这也是 pl 中要必须的格式： app_id_xxx。
			opts = append(opts, point.WithExtraTags(mq.Tagger.HostTags()))
		}
		if len(plMap) == 0 {
			err = fmt.Errorf("can not find [%s] pipeline script", topic)
			log.Warn(err)
			return err
		}
		fields["offset"] = msg.Offset
		fields["partition"] = msg.Partition
		pt := point.NewPointV2(topic, append(point.NewTags(tags), point.NewKVs(fields)...), opts...)

		if err := mq.feeder.FeedV2(category, []*point.Point{pt},
			dkio.WithInputName(topic),
			dkio.WithPipelineOption(&plmanager.Option{
				ScriptMap: plMap,
			}),
		); err != nil {
			log.Warnf("feed io err=%v", err)
		}
	}
	return err
}

func (mq *Custom) initMap() {
	if mq.LogTopicsMap == nil {
		mq.LogTopicsMap = make(map[string]string)
	}
	if mq.MetricTopicsMap == nil {
		mq.MetricTopicsMap = make(map[string]string)
	}
	if mq.RumTopicsMap == nil {
		mq.RumTopicsMap = make(map[string]string)
	}

	if mq.LogTopics != nil {
		for _, topic := range mq.LogTopics {
			mq.LogTopicsMap[topic] = mq.LogPl
		}
	}

	if mq.MetricTopics != nil {
		for _, topic := range mq.MetricTopics {
			mq.MetricTopicsMap[topic] = mq.MetricPl
		}
	}
}
