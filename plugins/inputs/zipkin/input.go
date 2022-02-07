// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

var (
	inputName = "zipkin"
	//nolint:lll
	sampleConfig = `
[[inputs.zipkin]]
  pathV1 = "/api/v1/spans"
  pathV2 = "/api/v2/spans"

  # [inputs.zipkin.tags]
    # tag1 = "value1"
    # tag2 = "value2"
    # ...
`
	tags map[string]string
	log  = logger.DefaultSLogger(inputName)
)

var (
	apiv1Path = "/api/v1/spans"
	apiv2Path = "/api/v2/spans"
)

type Input struct {
	PathV1 string            `toml:"pathV1"`
	PathV2 string            `toml:"pathV2"`
	Tags   map[string]string `toml:"tags"`
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

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if ipt.Tags != nil {
		tags = ipt.Tags
	}
}

func (ipt *Input) RegHTTPHandler() {
	itrace.StartTracingStatistic()

	if ipt.PathV1 == "" {
		ipt.PathV1 = apiv1Path
	}
	http.RegHTTPHandler("POST", ipt.PathV1, ZipkinTraceHandleV1)

	if ipt.PathV2 == "" {
		ipt.PathV2 = apiv2Path
	}
	http.RegHTTPHandler("POST", ipt.PathV2, ZipkinTraceHandleV2)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
