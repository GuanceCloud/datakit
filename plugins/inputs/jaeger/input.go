package jaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	inputName    = "jaeger"
	sampleConfig = `
[[inputs.jaeger]]
  # Jaeger endpoint for receiving tracing span over HTTP.
	# Default value set as below. DO NOT MODIFY THE ENDPOINT if not necessary.
  endpoint = "/apis/traces"

  # Jaeger agent host:port address for UDP transport.
  #	address = "127.0.0.1:6831"

  # [inputs.jaeger.tags]
    # tag1 = "val1"
    #	tag2 = "val2"
    # ...
`
	jaegerTags map[string]string
	log        = logger.DefaultSLogger(inputName)
)

type Input struct {
	Path             string                     `toml:"path"`           // deprecated
	UdpAgent         string                     `toml:"udp_agent"`      // deprecated
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"` // deprecated
	Endpoint         string                     `toml:"endpoint"`
	Address          string                     `toml:"address"`
	Tags             map[string]string          `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Tags != nil {
		jaegerTags = t.Tags
	}

	if t.Address != "" {
		StartUDPAgent(t.Address)
	}
}

func (t *Input) RegHttpHandler() {
	if t.Endpoint != "" {
		http.RegHttpHandler("POST", t.Endpoint, JaegerTraceHandle)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
