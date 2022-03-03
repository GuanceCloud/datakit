// Package jaeger handle Jaeger tracing metrics.
package jaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
	log                       = logger.DefaultSLogger(inputName)
	_          inputs.InputV2 = &Input{}
)

type Input struct {
	Path             string                     `toml:"path"`           // deprecated
	UDPAgent         string                     `toml:"udp_agent"`      // deprecated
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

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	// No predefined measurement available here
	return nil
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)
	iod.FeedEventLog(&iod.Reporter{Message: "jaeger start ok, ready for collecting metrics.", Logtype: "event"})

	if t.Tags != nil {
		jaegerTags = t.Tags
	}

	if t.Address != "" {
		if err := StartUDPAgent(t.Address); err != nil {
			log.Errorf("StartUDPAgent: %s", err)
		}
	}
}

func (t *Input) RegHTTPHandler() {
	if t.Endpoint != "" {
		http.RegHTTPHandler("POST", t.Endpoint, JaegerTraceHandle)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
