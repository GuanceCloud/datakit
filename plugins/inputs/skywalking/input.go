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

  ## Keep rare ddtrace resources list.
  # keep_rare_resource = false

  ## Ignore ddtrace resources list. List of strings
  ## A list of regular expressions used to block certain resource name.
  # [inputs.ddtrace.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config
  # [inputs.ddtrace.sampler]
    # priority = 0
    # sampling_rate = 1.0

  ## customer tags
  # [inputs.skywalking.V3.tags]
    # tag1 = "value1"
    # tag2 = "value2"
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
