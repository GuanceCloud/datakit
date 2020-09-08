package traceJaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceJaeger"

	traceJaegerConfigSample = `
#[inputs.traceJaeger]
#	path = "/api/traces"
#	[inputs.traceJaeger.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

var JaegerTags map[string]string

const (
	defaultJeagerPath = "/api/traces"
)

type JaegerTrace struct {
	Path string
	Tags map[string]string
}

func (_ *JaegerTrace) Catalog() string {
	return inputName
}

func (_ *JaegerTrace) SampleConfig() string {
	return traceJaegerConfigSample
}

func (t *JaegerTrace) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t != nil {
		JaegerTags = t.Tags
	}

	<-datakit.Exit.Wait()
	log.Infof("%s input exit", inputName)
}

func (t *JaegerTrace) RegHttpHandler() {
	if t.Path == "" {
		t.Path = defaultJeagerPath
	}
	http.RegHttpHandler("POST", t.Path, JaegerTraceHandle)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &JaegerTrace{}
		return t
	})
}
