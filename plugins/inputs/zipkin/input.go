// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceZipkin"
	//nolint:lll
	traceZipkinConfigSample = `
[[inputs.traceZipkin]]
  #	pathV1 = "/api/v1/spans"
  #	pathV2 = "/api/v2/spans"

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with first priority.
  # [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceZipkin.sample_configs.target]
    # env = "prod"

  ## Sub item in sample_configs list with second priority.
  # [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 100
    ## Sample scope, the range to be covered in once sample action.
    # scope = 1000
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceZipkin.sample_configs.target]
    # env = "dev"

    ## ...

  ## Sub item in sample_configs list with last priority.
  # [[inputs.traceZipkin.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched.
    # [inputs.traceZipkin.sample_configs.target]

  # [inputs.traceZipkin.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # ...
`
	zipkinTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

var (
	defaultZipkinPathV1  = "/api/v1/spans"
	defaultZipkinPathV2  = "/api/v2/spans"
	sampleConfs          []*trace.TraceSampleConfig
	zpkThriftV1Filters   []zipkinThriftV1SpansFilter
	zpkJSONV1Filters     []zipkinJSONV1SpansFilter
	zpkProtoBufV2Filters []zipkinProtoBufV2SpansFilter
	zpkJSONV2Filters     []zipkinJSONV2SpansFilter
)

type Input struct {
	PathV1           string                     `toml:"pathV1"`
	PathV2           string                     `toml:"pathV2"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	Tags             map[string]string          `toml:"tags"`
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

	sampleConfs = t.TraceSampleConfs
	for k, v := range sampleConfs {
		if v.Rate <= 0 || v.Scope < v.Rate {
			v.Rate = 100
			v.Scope = 100
			log.Warnf("%s input tracing sample config [%d] invalid, reset to default.", inputName, k)
		}
	}
	if len(sampleConfs) != 0 {
		zpkThriftV1Filters = append(zpkThriftV1Filters, zpkThriftV1Sample)
		zpkJSONV1Filters = append(zpkJSONV1Filters, zpkJSONV1Sample)
		zpkProtoBufV2Filters = append(zpkProtoBufV2Filters, zpkProtoBufV2Sample)
		zpkJSONV2Filters = append(zpkJSONV2Filters, zpkJSONV2Sample)
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
