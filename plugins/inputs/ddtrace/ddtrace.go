package ddtrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "ddtrace"

	traceDdtraceConfigSample = `
[inputs.ddtrace]
	# 此路由建议不要修改，以免跟其它路由冲突
	path = "/v0.4/traces"

	[inputs.ddtrace.tags]
		# tag1 = "tag1"
		# tag2 = "tag2"
		# tag3 = "tag3"
`
	log *logger.Logger
)

const (
	defaultDdtracePath = "/v0.4/traces"
)

var DdtraceTags map[string]string

type Zipkin struct {
	Tags map[string]string
}

type Ddtrace struct {
	Path string
	Tags map[string]string
}

func (_ *Ddtrace) Catalog() string {
	return inputName
}

func (_ *Ddtrace) SampleConfig() string {
	return traceDdtraceConfigSample
}

func (d *Ddtrace) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if d != nil {
		DdtraceTags = d.Tags
	}

	<-datakit.Exit.Wait()
	log.Infof("%s input exit", inputName)
}

func (d *Ddtrace) RegHttpHandler() {
	if d.Path == "" {
		d.Path = defaultDdtracePath
	}
	http.RegHttpHandler("POST", d.Path, DdtraceTraceHandle)
	http.RegHttpHandler("PUT", d.Path, DdtraceTraceHandle)
}
func (i *Ddtrace) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Ddtrace) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&DdtraceMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		d := &Ddtrace{}
		return d
	})
}
