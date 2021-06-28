package traceJaeger

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
	inputName               = "traceJaeger"
	traceJaegerConfigSample = `
[[inputs.traceJaeger]]
  #	path = "/api/traces"
  #	udp_agent = "127.0.0.1:6832"

  ## trace sample config, sample_rate and sample_scope together determine how many trace sample data will send to io
  [inputs.traceJaeger.sample_config]
    ## sample rate, how many will be sampled
    # rate = ` + fmt.Sprintf("%d", defRate) + `
    ## sample scope, the range to sample
    # scope = ` + fmt.Sprintf("%d", defScope) + `
    ## ignore tags list for samplingx
    # ignore_tags_list = []

  [inputs.traceJaeger.tags]
    # tag1 = "val1"
    #	tag2 = "val2"
    # tag3 = "val3"
`
	log = logger.DefaultSLogger(inputName)
)

var JaegerTags map[string]string

const (
	defaultJeagerPath = "/api/traces"
)

type JaegerTrace struct {
	Path            string                   `toml:"path"`
	UdpAgent        string                   `toml:"udp_agent"`
	TraceSampleConf *trace.TraceSampleConfig `toml:"sample_config"`
	Tags            map[string]string
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

	if t.TraceSampleConf != nil {
		if t.TraceSampleConf.Rate <= 0 {
			t.TraceSampleConf.Rate = defRate
		}
		if t.TraceSampleConf.Scope <= 0 {
			t.TraceSampleConf.Scope = defScope
		}
		traceSampleConf = t.TraceSampleConf
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
