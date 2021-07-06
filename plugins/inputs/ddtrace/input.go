package ddtrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

const (
	defaultDdtracePath = "/v0.4/traces"
)

var (
	defRate     = 15
	defScope    = 100
	sampleConfs []*trace.TraceSampleConfig
)

var (
	inputName                = "ddtrace"
	traceDdtraceConfigSample = `
[[inputs.ddtrace]]
  # 此路由建议不要修改，以免跟其它路由冲突
  path = "/v0.4/traces"

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with priority 1.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    # [inputs.ddtrace.sample_configs.target]
    # tag = "value"

  ## Sub item in sample_configs list with priority 2.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    # [inputs.ddtrace.sample_configs.target]
    # tag = "value"

  ## ...

  ## Sub item in sample_configs list with priority n.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched, so that this pair shoud be empty.
    # [inputs.ddtrace.sample_configs.target]

	## customer tags
  # [inputs.ddtrace.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    ## ...
`
	DdtraceTags map[string]string
	log         = logger.DefaultSLogger(inputName)
)

type Input struct {
	Path             string                     `toml:"path"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	Tags             map[string]string          `toml:"tags"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return traceDdtraceConfigSample
}

func (d *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	sampleConfs = d.TraceSampleConfs
	// check tracing sample config
	for k, v := range sampleConfs {
		if v.Rate <= 0 || v.Scope < v.Rate {
			v.Rate = 100
			v.Scope = 100
			log.Warnf("%s input tracing sample config [%d] invalid, reset to default.", inputName, k)
		}
	}

	if d.Tags != nil {
		DdtraceTags = d.Tags
	}

	if d != nil {
		<-datakit.Exit.Wait()
		log.Infof("%s input exit", inputName)
	}
}

func (d *Input) RegHttpHandler() {
	if d.Path == "" {
		d.Path = defaultDdtracePath
	}
	http.RegHttpHandler("POST", d.Path, DdtraceTraceHandle)
	http.RegHttpHandler("PUT", d.Path, DdtraceTraceHandle)
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&DdtraceMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
