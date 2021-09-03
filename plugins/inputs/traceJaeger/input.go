package traceJaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	inputName               = "traceJaeger"
	traceJaegerConfigSample = `
[[inputs.traceJaeger]]
  #	path = "/api/traces"
  #	udp_agent = "127.0.0.1:6832"

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with first priority.
  # [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceJaeger.sample_configs.target]
    # env = "prod"

  ## Sub item in sample_configs list with second priority.
  # [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 100
    ## Sample scope, the range to be covered in once sample action.
    # scope = 1000
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceJaeger.sample_configs.target]
    # env = "dev"

    ## ...

  ## Sub item in sample_configs list with last priority.
  # [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched.
    # [inputs.traceJaeger.sample_configs.target]

  # [inputs.traceJaeger.tags]
    # tag1 = "val1"
    #	tag2 = "val2"
    # ...
`
	jaegerTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

var (
	defaultJeagerPath = "/api/traces"
	sampleConfs       []*trace.TraceSampleConfig
	filters           []batchFilter
)

type Input struct {
	Path             string                     `toml:"path"`
	UdpAgent         string                     `toml:"udp_agent"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	Tags             map[string]string          `toml:"tags"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return traceJaegerConfigSample
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Tags != nil {
		jaegerTags = t.Tags
	}

	if t.UdpAgent != "" {
		StartUdpAgent(t.UdpAgent)
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
		filters = append(filters, sample)
	}
}

func (t *Input) RegHttpHandler() {
	if t.Path == "" {
		t.Path = defaultJeagerPath
	}
	http.RegHttpHandler("POST", t.Path, JaegerTraceHandle)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
