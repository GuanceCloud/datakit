package traceZipkin

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

const (
	defaultZipkinPathV1 = "/api/v1/spans"
	defaultZipkinPathV2 = "/api/v2/spans"
)

var (
	defRate         = 15
	defScope        = 100
	traceSampleConf *trace.TraceSampleConfig
)

var (
	inputName               = "traceZipkin"
	traceZipkinConfigSample = `
[[inputs.traceZipkin]]
  #	pathV1 = "/api/v1/spans"
  #	pathV2 = "/api/v2/spans"

  ## trace sample config, sample_rate and sample_scope together determine how many trace sample data will send to io
  # [inputs.traceZipkin.sample_config]
    ## sample rate, how many will be sampled
    # rate = ` + fmt.Sprintf("%d", defRate) + `
    ## sample scope, the range to sample
    # scope = ` + fmt.Sprintf("%d", defScope) + `
    ## ignore tags list for samplingx
    # ignore_tags_list = []

  # [inputs.traceZipkin.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # ...
`
	ZipkinTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

type Input struct {
	PathV1          string                   `toml:"pathV1"`
	PathV2          string                   `toml:"pathV2"`
	TraceSampleConf *trace.TraceSampleConfig `toml:"sample_config"`
	Tags            map[string]string        `toml:"tags"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return traceZipkinConfigSample
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Tags != nil {
		ZipkinTags = t.Tags
	}

	if t.TraceSampleConf != nil {
		if t.TraceSampleConf.Rate <= 0 {
			t.TraceSampleConf.Rate = defRate
		}
		if t.TraceSampleConf.Scope <= 0 {
			t.TraceSampleConf.Scope = defScope
		}
		traceSampleConf = t.TraceSampleConf
	}

	<-datakit.Exit.Wait()

	log.Infof("%s input exit", inputName)
}

func (t *Input) RegHttpHandler() {
	if t.PathV1 == "" {
		t.PathV1 = defaultZipkinPathV1
	}
	http.RegHttpHandler("POST", t.PathV1, ZipkinTraceHandleV1)

	if t.PathV2 == "" {
		t.PathV2 = defaultZipkinPathV2
	}
	http.RegHttpHandler("POST", t.PathV2, ZipkinTraceHandleV2)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
