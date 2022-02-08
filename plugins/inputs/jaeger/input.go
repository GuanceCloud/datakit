// Package jaeger handle Jaeger tracing metrics.
package jaeger

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

var (
	inputName    = "jaeger"
	sampleConfig = `
[[inputs.jaeger]]
  # Jaeger endpoint for receiving tracing span over HTTP.
  # Default value set as below. DO NOT MODIFY THE ENDPOINT if not necessary.
  endpoint = "/apis/traces"

  # Jaeger agent host:port address for UDP transport.
  # address = "127.0.0.1:6831"

  # [inputs.jaeger.tags]
    # tag1 = "value1"
    # tag2 = "value2"
    # ...
`
	tags = make(map[string]string)
	log  = logger.DefaultSLogger(inputName)
)

var afterGather = itrace.NewAfterGather()

type Input struct {
	Path     string            `toml:"path"`      // deprecated
	UDPAgent string            `toml:"udp_agent"` // deprecated
	Endpoint string            `toml:"endpoint"`
	Address  string            `toml:"address"`
	Tags     map[string]string `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)
	dkio.FeedEventLog(&dkio.Reporter{Message: "jaeger start ok, ready for collecting metrics.", Logtype: "event"})

	// add calculators
	afterGather.AppendCalculator(itrace.StatTracingInfo)

	// start up UDP agent
	if ipt.Address != "" {
		itrace.StartTracingStatistic()
		if err := StartUDPAgent(ipt.Address); err != nil {
			log.Errorf("%s start UDP agent failed: %s", inputName, err.Error())
		}
	}

	if len(ipt.Tags) != 0 {
		tags = ipt.Tags
	}
}

func (ipt *Input) RegHTTPHandler() {
	if ipt.Endpoint != "" {
		itrace.StartTracingStatistic()
		http.RegHTTPHandler("POST", ipt.Endpoint, JaegerTraceHandle)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
