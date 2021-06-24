package ddtrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "ddtrace"

	traceDdtraceConfigSample = `
[inputs.ddtrace]
  # 此路由建议不要修改，以免跟其它路由冲突
  path = "/v0.4/traces"

  ## trace input sample status: true: on, false: off
  # sample_status = false
  ## sample config
  # [inputs.ddtrace.sample_config]
    ## sample rate, rate and scope determine rate number of input out of scope number will be sampled
    # rate = 15
    ## sample scope
    # scope = 100
    # [inputs.ddtrace.sample_config.ignore_list]
      # key1 = "value1"
      # key2 = "value2"
      ## ...

  [inputs.ddtrace.tags]
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
	Path            string                `toml:"path"`
	TraceSampleConf *io.TraceSampleConfig `toml:"sample_config"`
	Tags            map[string]string     `toml:"tags"`
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
		d := &Input{}
		d.TraceSampleConf.SampleHandler = io.DefSampleHandler
		d.TraceSampleConf.Key = "trace_id"
		d.TraceSampleConf.ConvertHandler = io.DefConvertHandler

		return d
	})
}
