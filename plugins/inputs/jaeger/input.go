// Package jaeger handle Jaeger tracing metrics.
package jaeger

import (
	"time"

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

const (
	inputName    = "jaeger"
	sampleConfig = `
[[inputs.jaeger]]
  # Jaeger endpoint for receiving tracing span over HTTP.
  # Default value set as below. DO NOT MODIFY THE ENDPOINT if not necessary.
  endpoint = "/apis/traces"

  # Jaeger agent host:port address for UDP transport.
  # address = "127.0.0.1:6831"

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.jaeger.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  # [inputs.jaeger.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##   -1: always reject any tracing data send to datakit
  ##    0: accept tracing data and calculate with sampling_rate
  ##    1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.jaeger.sampler]
    # priority = 0
    # sampling_rate = 1.0

  # [inputs.jaeger.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...
`
)

var (
	log              = logger.DefaultSLogger(inputName)
	afterGather      = itrace.NewAfterGather()
	keepRareResource *itrace.KeepRareResource
	closeResource    *itrace.CloseResource
	defSampler       *itrace.Sampler
	customerKeys     []string
	tags             map[string]string
)

type Input struct {
	Path             string              `toml:"path"`      // deprecated
	UDPAgent         string              `toml:"udp_agent"` // deprecated
	Endpoint         string              `toml:"endpoint"`
	Address          string              `toml:"address"`
	CustomerTags     []string            `toml:"customer_tags"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	Tags             map[string]string   `toml:"tags"`
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

	// add filters: the order append in AfterGather is important!!!
	// add close resource filter
	if len(ipt.CloseResource) != 0 {
		closeResource = &itrace.CloseResource{}
		closeResource.UpdateIgnResList(ipt.CloseResource)
		afterGather.AppendFilter(closeResource.Close)
	}
	// add rare resource keeper
	if ipt.KeepRareResource {
		keepRareResource = &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add sampler
	if ipt.Sampler != nil {
		defSampler = ipt.Sampler
		afterGather.AppendFilter(defSampler.Sample)
	}

	// start up UDP agent
	if ipt.Address != "" {
		itrace.StartTracingStatistic()
		if err := StartUDPAgent(ipt.Address); err != nil {
			log.Errorf("%s start UDP agent failed: %s", inputName, err.Error())
		}
	}

	customerKeys = ipt.CustomerTags
	tags = ipt.Tags
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
