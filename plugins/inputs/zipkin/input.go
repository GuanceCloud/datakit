// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceZipkin"
	//nolint:lll
	traceZipkinConfigSample = `
[[inputs.traceZipkin]]
  pathV1 = "/api/v1/spans"
  pathV2 = "/api/v2/spans"

  # [inputs.traceZipkin.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # ...
`
	zipkinTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

var (
	defaultZipkinPathV1 = "/api/v1/spans"
	defaultZipkinPathV2 = "/api/v2/spans"
)

type Input struct {
	PathV1 string `toml:"pathV1"`
	PathV2 string `toml:"pathV2"`
	// TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	Tags map[string]string `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return traceZipkinConfigSample
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Tags != nil {
		zipkinTags = t.Tags
	}
}

func (t *Input) RegHTTPHandler() {
	if t.PathV1 == "" {
		t.PathV1 = defaultZipkinPathV1
	}
	http.RegHTTPHandler("POST", t.PathV1, ZipkinTraceHandleV1)

	if t.PathV2 == "" {
		t.PathV2 = defaultZipkinPathV2
	}
	http.RegHTTPHandler("POST", t.PathV2, ZipkinTraceHandleV2)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
