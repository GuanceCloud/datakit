package ddtrace

import (
	"strings"

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
	inputName           = "ddtrace"
	ddtraceSampleConfig = `
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
    # env = "prod"

  ## Sub item in sample_configs list with priority 2.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 100
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 1000
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    # [inputs.ddtrace.sample_configs.target]
    # env = "dev"

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

  ## Customer tag prefix used in client code like span.SetSpan([customer_tag_prefix]key, value)
  ## ddtrace collector will not trim the prefix in order to avoid tags confliction. IT'S EMPTY STRING VALUE AS DEFAULT
  ## indicates that no customer tag set up. DO NOT USE DOT(.) IN
  # customer_tag_prefix = ""

  ## customer tags
  # [inputs.ddtrace.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    ## ...
`
	DdtraceTags         map[string]string
	customerTagPrefixes = map[string]string{}
	log                 = logger.DefaultSLogger(inputName)
)

type Input struct {
	Path              string                     `toml:"path"`
	TraceSampleConfs  []*trace.TraceSampleConfig `toml:"sample_configs"`
	CustomerTagPrefix string                     `toml:"customer_tag_prefix"`
	Tags              map[string]string          `toml:"tags"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return ddtraceSampleConfig
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&DdtraceMeasurement{},
	}
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

	if d.CustomerTagPrefix != "" {
		if strings.Contains(d.CustomerTagPrefix, ".") {
			d.CustomerTagPrefix = strings.ReplaceAll(d.CustomerTagPrefix, ".", "_")
		}
		customerTagPrefixes[d.Path] = d.CustomerTagPrefix
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

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
