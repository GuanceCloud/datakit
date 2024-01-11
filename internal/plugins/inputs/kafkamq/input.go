// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/custom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/handle"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/jaeger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/opentelemetry"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/skywalking"
)

var _ inputs.InputV2 = &Input{}

const mqSampleConfig = `
[[inputs.kafkamq]]
  addrs = ["localhost:9092"]
  # your kafka version:0.8.2 ~ 3.2.0
  kafka_version = "2.0.0"
  group_id = "datakit-group"
  # consumer group partition assignment strategy (range, roundrobin, sticky)
  assignor = "roundrobin"

  ## rate limit.
  #limit_sec = 100
  ## sample
  # sampling_rate = 1.0

  ## kafka tls config
  # tls_enable = true
  # tls_security_protocol = "SASL_PLAINTEXT"
  # tls_sasl_mechanism = "PLAIN"
  # tls_sasl_plain_username = "user"
  # tls_sasl_plain_password = "pw"

  ## -1:Offset Newest, -2:Offset Oldest
  offsets=-1

  ## skywalking custom
  #[inputs.kafkamq.skywalking]
    ## Required！send to datakit skywalking input.
    #dk_endpoint="http://localhost:9529"
    #thread = 8 
    #topics = [
    #  "skywalking-metrics",
    #  "skywalking-profilings",
    #  "skywalking-segments",
    #  "skywalking-managements",
    #  "skywalking-meters",
    #  "skywalking-logging",
    #]
    #namespace = ""

  ## Jaeger from kafka. Please make sure your Datakit Jaeger collector is open ！！！
  #[inputs.kafkamq.jaeger]
    ## Required！ ipv6 is "[::1]:9529"
    #dk_endpoint="http://localhost:9529"
    #thread = 8 
    #source: agent,otel,others...
    #source = "agent"
    ## Required！ topics
    #topics=["jaeger-spans","jaeger-my-spans"]

  ## user custom message with PL script.
  #[inputs.kafkamq.custom]
    #spilt_json_body = true
    #thread = 8 
    ## spilt_topic_map determines whether to enable log splitting for specific topic based on the values in the spilt_topic_map[topic].
    #[inputs.kafkamq.custom.spilt_topic_map]
    #  "log_topic"=true
    #  "log01"=false
    #[inputs.kafkamq.custom.log_topic_map]
    #  "log_topic"="log.p"
    #  "log01"="log_01.p"
    #[inputs.kafkamq.custom.metric_topic_map]
    #  "metric_topic"="metric.p"
    #  "metric01"="rum_apm.p"
    #[inputs.kafkamq.custom.rum_topic_map]
    #  "rum_topic"="rum_01.p"
    #  "rum_02"="rum_02.p"

  #[inputs.kafkamq.remote_handle]
    ## Required！
    #endpoint="http://localhost:8080"
    ## Required！ topics
    #topics=["spans","my-spans"]
    # send_message_count = 100
    # debug = false
    # is_response_point = true
    # header_check = false
  
  ## Receive and consume OTEL data from kafka.
  #[inputs.kafkamq.otel]
    #dk_endpoint="http://localhost:9529"
    #trace_api="/otel/v1/trace"
    #metric_api="/otel/v1/metric"
    #trace_topics=["trace1","trace2"]
    #metric_topics=["otel-metric","otel-metric1"]
    #thread = 8 

  ## todo: add other input-mq
`

var (
	log       = logger.DefaultSLogger(inputName)
	inputName = "kafkamq"
)

type Input struct {
	Addr                 string   `toml:"addr"`          // 1.4 开始废弃
	Addrs                []string `toml:"addrs"`         // 向下兼容 addr
	KafkaVersion         string   `toml:"kafka_version"` // kafka version
	GroupID              string   `toml:"group_id"`      // 消费组 ID
	Assignor             string   `toml:"assignor"`      // 消费模式
	LimitSec             int      `toml:"limit_sec"`     // 令牌桶速率限制，每秒多少个.
	SamplingRate         float64  `toml:"sampling_rate"` // 采集率：主要针对自定义topic
	TLSEnable            bool     `toml:"tls_enable"`
	TLSSecurityProtocol  string   `toml:"tls_security_protocol"`
	TLSSaslMechanism     string   `toml:"tls_sasl_mechanism"`
	TLSSaslPlainUsername string   `toml:"tls_sasl_plain_username"`
	TLSSaslPlainPassword string   `toml:"tls_sasl_plain_password"`
	Offsets              int64    `toml:"offsets"`

	SkyWalking *skywalking.SkyConsumer   `toml:"skywalking"`    // 命名时 注意区分源
	Jaeger     *jaeger.Consumer          `toml:"jaeger"`        // 命名时 注意区分源
	Custom     *custom.Custom            `toml:"custom"`        // 自定义 topic
	Handle     *handle.Handle            `toml:"remote_handle"` // 自定义 handle
	OTELHandle *opentelemetry.OTELHandle `toml:"otel"`          // 自定义 handle

	kafka  *kafkaConsumer
	feeder dkio.Feeder
}

func (*Input) Catalog() string      { return "kafkamq" }
func (*Input) SampleConfig() string { return mqSampleConfig }

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("init input = %+v", ipt)
	if ipt.feeder == nil {
		log.Infof("feeder is nil, use default Feeder")
		ipt.feeder = dkio.DefaultFeeder()
	}

	addrs := getAddrs(ipt.Addr, ipt.Addrs)
	config := newSaramaConfig(withVersion(ipt.KafkaVersion),
		withAssignors(ipt.Assignor),
		withOffset(ipt.Offsets),
		withSASL(ipt.TLSEnable, ipt.TLSSaslMechanism, ipt.TLSSaslPlainUsername, ipt.TLSSaslPlainPassword),
	)

	ipt.kafka = &kafkaConsumer{
		process: make(map[string]TopicProcess),
		topics:  make([]string, 0),
		addrs:   addrs,
		groupID: ipt.GroupID,
		stop:    make(chan struct{}),
		config:  config,
		ready:   make(chan bool),
	}
	ipt.kafka.limitAndSample(ipt.LimitSec, ipt.SamplingRate)

	if ipt.SkyWalking != nil {
		ipt.kafka.registerP(ipt.SkyWalking)
	}

	if ipt.Custom != nil {
		ipt.Custom.SetFeeder(ipt.feeder)
		ipt.kafka.registerP(ipt.Custom)
	}

	if ipt.Jaeger != nil {
		ipt.Jaeger.SetFeeder(ipt.feeder)
		ipt.kafka.registerP(ipt.Jaeger)
	}

	if ipt.Handle != nil {
		ipt.Handle.SetFeeder(ipt.feeder)
		ipt.kafka.registerP(ipt.Handle)
	}

	if ipt.OTELHandle != nil {
		ipt.kafka.registerP(ipt.OTELHandle)
	}

	go ipt.kafka.start()
}

func (ipt *Input) Terminate() {
	if ipt.kafka.stop != nil {
		close(ipt.kafka.stop)
	}
	log.Infof("input[%s] exit", inputName)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
