// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/skywalkingapi"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kafkamq/skywalking"
)

const mqSampleConfig = `
[[inputs.kafkamq]]
  addrs = ["localhost:9092"]
  # your kafka version:0.8.2.0 ~ 2.8.0.0
  kafka_version = "2.8.0.0"
  group_id = "datakit-group"
  plugins = ["db.type"]
  # Consumer group partition assignment strategy (range, roundrobin, sticky)
  assignor = "roundrobin"

  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  [inputs.kafkamq.skywalking]  
    topics = [
      "skywalking-metrics",
      "skywalking-profilings",
      "skywalking-segments",
      "skywalking-managements",
      "skywalking-meters",
      "skywalking-logging",
    ]
    namespace = ""

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.kafkamq.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.kafkamq.sampler]
    # sampling_rate = 1.0

  # [inputs.kafkamq.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.kafkamq.storage]
    # path = "./skywalking_storage"
    # capacity = 5120

  ## todo: add other input-mq 
`

var (
	_         inputs.InputV2 = &Input{}
	log                      = logger.DefaultSLogger(inputName)
	inputName                = "kafkamq"
)

type Input struct {
	Addr         string                  `toml:"addr"`
	Addrs        []string                `toml:"addrs"`         // 向下兼容 addr
	KafkaVersion string                  `toml:"kafka_version"` // kafka version
	GroupID      string                  `toml:"group_id"`
	SkyWalking   *skywalking.SkyConsumer `toml:"skywalking"` // 命名时 注意区分源
	Assignor     string                  `toml:"assignor"`   // 消费模式

	Plugins          []string            `toml:"plugins"`
	CustomerTags     []string            `toml:"customer_tags"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	Tags             map[string]string   `toml:"tags"`
	Storage          *itrace.Storage     `toml:"storage"`
}

func (*Input) Catalog() string      { return "kafkamq" }
func (*Input) SampleConfig() string { return mqSampleConfig }

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (i *Input) Terminate() {
	if i.SkyWalking != nil {
		i.SkyWalking.Stop()
	}
	log.Infof("input[%s] exit", inputName)
}

func (i Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&skywalkingapi.MetricMeasurement{},
	}
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("init input = %v", i)

	api := skywalkingapi.InitApiPluginAges(i.Plugins, i.Storage, i.CloseResource, i.KeepRareResource, i.Sampler, i.CustomerTags, i.Tags, inputName)
	addrs := getAddrs(i.Addr, i.Addrs)
	version := getKafkaVersion(i.KafkaVersion)
	balance := getAssignors(i.Assignor)
	if i.SkyWalking != nil {
		g := goroutine.NewGroup(goroutine.Option{Name: "inputs_kafkamq"})
		g.Go(func(ctx context.Context) error {
			log.Infof("start input kafkamq")
			i.SkyWalking.SaramaConsumerGroup(addrs, i.GroupID, api, version, balance)
			return nil
		})
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
