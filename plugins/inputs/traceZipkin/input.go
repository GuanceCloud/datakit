package traceZipkin

import (
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
	sampleConfs []*trace.TraceSampleConfig
)

var (
	inputName               = "traceZipkin"
	traceZipkinConfigSample = `
[[inputs.traceZipkin]]
  #	pathV1 = "/api/v1/spans"
  #	pathV2 = "/api/v2/spans"

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with priority 1.
  [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    [inputs.traceZipkin.sample_configs.target]
    tag = "value"

  ## Sub item in sample_configs list with priority 2.
  [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    [inputs.traceZipkin.sample_configs.target]
    tag = "value"

  ## ...

  ## Sub item in sample_configs list with priority n.
  [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched.
    # [inputs.traceZipkin.sample_configs.target]
    # tag = "value"

  # [inputs.traceZipkin.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # ...
`
	ZipkinTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

type Input struct {
	PathV1           string                     `toml:"pathV1"`
	PathV2           string                     `toml:"pathV2"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	Tags             map[string]string          `toml:"tags"`
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

	sampleConfs = t.TraceSampleConfs
	// check tracing sample config
	for k, v := range sampleConfs {
		if v.Rate <= 0 || v.Scope < v.Rate {
			v.Rate = 100
			v.Scope = 100
			log.Warnf("%s input tracing sample config [%d] invalid, reset to default.", inputName, k)
		}
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
