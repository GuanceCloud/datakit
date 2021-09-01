package traceJaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

const (
	defaultJeagerPath = "/api/traces"
)

var (
	sampleConfs []*trace.TraceSampleConfig
)

var (
	inputName               = "traceJaeger"
	traceJaegerConfigSample = `
[[inputs.traceJaeger]]
  #	path = "/api/traces"
  #	udp_agent = "127.0.0.1:6832"

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with priority 1.
  [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    [inputs.traceJaeger.sample_configs.target]
    tag = "value"

  ## Sub item in sample_configs list with priority 2.
  [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    [inputs.traceJaeger.sample_configs.target]
    tag = "value"

  ## ...

  ## Sub item in sample_configs list with priority n.
  [[inputs.traceJaeger.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched.
    # [inputs.traceJaeger.sample_configs.target]
    # tag = "value"

  # [inputs.traceJaeger.tags]
    # tag1 = "val1"
    #	tag2 = "val2"
    # ...
`
	JaegerTags map[string]string
	log        = logger.DefaultSLogger(inputName)
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
		JaegerTags = t.Tags
	}

	if t.UdpAgent != "" {
		StartUdpAgent(t.UdpAgent)
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
