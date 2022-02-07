// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
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
  ## customer tags
  # [inputs.skywalking.V3.tags]
    # tag1 = "value1"
    # tag2 = "value2"
    # ...
`
	defAddr = "localhost:13800"
	tags    map[string]string
	log     = logger.DefaultSLogger(inputName)
)

// deprecated.
type skywalkingConfig struct {
	Address string            `toml:"address"`
	Tags    map[string]string `toml:"tags"`
}

type Input struct {
	V2      *skywalkingConfig `toml:"V2"` // deprecated
	V3      *skywalkingConfig `toml:"V3"` // deprecated
	Address string            `toml:"address"`
	Tags    map[string]string `toml:"tags"`
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
