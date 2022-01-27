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
	configSample = `
[[inputs.zipkin]]
  pathV1 = "/api/v1/spans"
  pathV2 = "/api/v2/spans"

  # [inputs.zipkin.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # ...
`
	zipkinTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

var (
	pathV1 = "/api/v1/spans"
	pathV2 = "/api/v2/spans"
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
	return configSample
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Tags != nil {
		zipkinTags = t.Tags
	}
}

func (t *Input) RegHTTPHandler() {
	itrace.StartTracingStatistic()

	if t.PathV1 == "" {
		t.PathV1 = pathV1
	}
	http.RegHTTPHandler("POST", t.PathV1, ZipkinTraceHandleV1)

	if t.PathV2 == "" {
		t.PathV2 = pathV2
	}
	http.RegHTTPHandler("POST", t.PathV2, ZipkinTraceHandleV2)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
