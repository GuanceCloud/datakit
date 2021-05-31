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
[inputs.traceJaeger]
#	path = "/api/traces"
#	udp_agent = "127.0.0.1:6832"
	[inputs.traceJaeger.tags]
#		tag1 = "val1"
#		tag2 = "val2"
#		tag3 = "val3"
`
	log *logger.Logger
)

var JaegerTags map[string]string

const (
	defaultJeagerPath = "/api/traces"
)

type JaegerTrace struct {
	Path     string
	UdpAgent string `toml:"udp_agent"`
	Tags     map[string]string
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

	if t.UdpAgent != "" {
		StartUdpAgent(t.UdpAgent)
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
