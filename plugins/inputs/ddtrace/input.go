package ddtrace

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	defRate         = 15
	defScope        = 100
	traceSampleConf *trace.TraceSampleConfig
)

var (
	inputName                = "ddtrace"
	traceDdtraceConfigSample = `
[[inputs.ddtrace]]
  # 此路由建议不要修改，以免跟其它路由冲突
  path = "/v0.4/traces"

  ## trace sample config, sample_rate and sample_scope together determine how many trace sample data will send to io
  # [inputs.ddtrace.sample_config]
    ## sample rate, how many will be sampled
    # rate = ` + fmt.Sprintf("%d", defRate) + `
    ## sample scope, the range to sample
    # scope = ` + fmt.Sprintf("%d", defScope) + `
    ## ignore tags list for samplingx
    # ignore_tags_list = []

  # [inputs.ddtrace.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    ## ...
`
	log = logger.DefaultSLogger(inputName)
)

const (
	defaultDdtracePath = "/v0.4/traces"
)

var DdtraceTags map[string]string

type Input struct {
	Path            string                   `toml:"path"`
	TraceSampleConf *trace.TraceSampleConfig `toml:"sample_config"`
	Tags            map[string]string        `toml:"tags"`
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

	if d.TraceSampleConf != nil {
		if d.TraceSampleConf.Rate <= 0 {
			d.TraceSampleConf.Rate = defRate
		}
		if d.TraceSampleConf.Scope <= 0 {
			d.TraceSampleConf.Scope = defScope
		}
		traceSampleConf = d.TraceSampleConf
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
