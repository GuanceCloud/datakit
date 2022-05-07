// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.InputV2 = &Input{}

const (
	inputName    = "skywalking"
	sampleConfig = `
[[inputs.skywalking]]
  ## skywalking grpc server listening on address
  address = "localhost:13800"

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.skywalking.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  # [inputs.skywalking.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##  -1: always reject any tracing data send to datakit
  ##   0: accept tracing data and calculate with sampling_rate
  ##   1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.skywalking.sampler]
    # priority = 0
    # sampling_rate = 1.0

  ## Piplines use to manipulate message and meta data. If this item configured right then
  ## the current input procedure will run the scripts wrote in pipline config file against the data
  ## present in span message.
  ## The string on the left side of the equal sign must be identical to the service name that
  ## you try to handle.
  # [inputs.skywalking.pipelines]
    # service1 = "service1.p"
    # service2 = "service2.p"
    # ...

  # [inputs.skywalking.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...
`
)

var (
	log              = logger.DefaultSLogger(inputName)
	defAddr          = "localhost:13800"
	afterGatherRun   itrace.AfterGatherHandler
	keepRareResource *itrace.KeepRareResource
	closeResource    *itrace.CloseResource
	sampler          *itrace.Sampler
	customerKeys     []string
	tags             map[string]string
)

type Input struct {
	V2               interface{}         `toml:"V2"` // deprecated *skywalkingConfig
	V3               interface{}         `toml:"V3"` // deprecated *skywalkingConfig
	Address          string              `toml:"address"`
	CustomerTags     []string            `toml:"customer_tags"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	Pipelines        map[string]string   `toml:"pipelines"`
	Tags             map[string]string   `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if len(ipt.Address) == 0 {
		ipt.Address = defAddr
	}

	afterGather := itrace.NewAfterGather()
	afterGatherRun = afterGather

	// add calculators
	// afterGather.AppendCalculator(itrace.StatTracingInfo)

	// add filters: the order append in AfterGather is important!!!
	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add close resource filter
	if len(ipt.CloseResource) != 0 {
		closeResource = &itrace.CloseResource{}
		closeResource.UpdateIgnResList(ipt.CloseResource)
		afterGather.AppendFilter(closeResource.Close)
	}
	// add rare resource keeper
	if ipt.KeepRareResource {
		keepRareResource = &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add sampler
	if ipt.Sampler != nil {
		sampler = ipt.Sampler
		afterGather.AppendFilter(sampler.Sample)
	}
	// add piplines
	if len(ipt.Pipelines) != 0 {
		afterGather.AppendFilter(itrace.PiplineFilterWrapper(inputName, ipt.Pipelines))
	}

	customerKeys = ipt.CustomerTags
	tags = ipt.Tags

	log.Debug("start skywalking grpc v3 server")

	// itrace.StartTracingStatistic()
	go registerServerV3(ipt.Address)
}

func (ipt *Input) Terminate() {
	// TODO: 必须写
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
