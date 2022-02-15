// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.InputV2 = &Input{}

var (
	inputName    = "skywalking"
	sampleConfig = `
[[inputs.skywalking]]
  ## skywalking grpc server listening on address
  address = "localhost:13800"

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.skywalking.tags]. DO NOT CONTAIN DOT(.) IN KEYS LIST.
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
  ##   -1: always reject any tracing data send to datakit
  ##    0: accept tracing data and calculate with sampling_rate
  ##    1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.skywalking.sampler]
    # priority = 0
    # sampling_rate = 1.0

  # [inputs.skywalking.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...
`
	defAddr = "localhost:13800"
	tags    = make(map[string]string)
	log     = logger.DefaultSLogger(inputName)
)

var (
	afterGather      = itrace.NewAfterGather()
	keepRareResource *itrace.KeepRareResource
	closeResource    *itrace.CloseResource
	defSampler       *itrace.Sampler
)

type Input struct {
	V2               interface{}         `toml:"V2"` // deprecated *skywalkingConfig
	V3               interface{}         `toml:"V3"` // deprecated *skywalkingConfig
	Address          string              `toml:"address"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
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

	// add calculators
	afterGather.AppendCalculator(itrace.StatTracingInfo)

	// add filters: the order append in AfterGather is important!!!
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
		defSampler = ipt.Sampler
		afterGather.AppendFilter(defSampler.Sample)
	}

	if len(ipt.Tags) != 0 {
		tags = ipt.Tags
	}

	log.Debug("start skywalking grpc v3 server")

	itrace.StartTracingStatistic()
	go runServerV3(ipt.Address)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
