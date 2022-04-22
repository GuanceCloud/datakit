// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "ddtrace"
	sampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.ddtrace.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  # [inputs.ddtrace.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##  -1: always reject any tracing data send to datakit
  ##   0: accept tracing data and calculate with sampling_rate
  ##   1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.ddtrace.sampler]
    # priority = 0
    # sampling_rate = 1.0

  ## Piplines use to manipulate message and meta data. If this item configured right then
  ## the current input procedure will run the scripts wrote in pipline config file against the data
  ## present in span message.
  ## The string on the left side of the equal sign must be identical to the service name that
  ## you try to handle.
  # [inputs.ddtrace.pipelines]
    # service1 = "service1.p"
    # service2 = "service2.p"
    # ...

  # [inputs.ddtrace.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...
`
)

var (
	log                                          = logger.DefaultSLogger(inputName)
	v1, v2, v3, v4, v5                           = "/v0.1/spans", "/v0.2/traces", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces"
	info, stats                                  = "/info", "/v0.6/stats"
	afterGather                                  = itrace.NewAfterGather()
	afterGatherRun     itrace.AfterGatherHandler = afterGather
	keepRareResource   *itrace.KeepRareResource
	closeResource      *itrace.CloseResource
	sampler            *itrace.Sampler
	customerKeys       = []string{"runtime-id"}
	tags               map[string]string
)

type Input struct {
	Path             string              `toml:"path,omitempty"`           // deprecated
	TraceSampleConfs interface{}         `toml:"sample_configs,omitempty"` // deprecated []*itrace.TraceSampleConfig
	TraceSampleConf  interface{}         `toml:"sample_config"`            // deprecated *itrace.TraceSampleConfig
	IgnoreResources  []string            `toml:"ignore_resources"`         // deprecated []string
	Endpoints        []string            `toml:"endpoints"`
	CustomerTags     []string            `toml:"customer_tags"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	Pipelines        map[string]string   `toml:"pipelines"`
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

func (ipt *Input) RegHTTPHandler() {
	var isReg bool
	for _, endpoint := range ipt.Endpoints {
		switch endpoint {
		case v1, v2, v3, v4, v5:
			isReg = true
			dkhttp.RegHTTPHandler(http.MethodPost, endpoint, handleDDTraceWithVersion(endpoint))
			dkhttp.RegHTTPHandler(http.MethodPut, endpoint, handleDDTraceWithVersion(endpoint))
			log.Infof("pattern %s registered", endpoint)
		default:
			log.Errorf("unrecognized ddtrace agent endpoint")
		}
	}
	if isReg {
		// itrace.StartTracingStatistic()
		// unsupported api yet
		dkhttp.RegHTTPHandler(http.MethodPost, info, handleDDInfo)
		dkhttp.RegHTTPHandler(http.MethodPost, stats, handleDDStats)
	}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	// add calculators
	// afterGather.AppendCalculator(itrace.StatTracingInfo)

	// add filters: the order append in AfterGather is important!!!
	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
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
		sampler = ipt.Sampler
		afterGather.AppendFilter(sampler.Sample)
	}
	// add piplines
	if len(ipt.Pipelines) != 0 {
		afterGather.AppendFilter(itrace.PiplineFilterWrapper(inputName, ipt.Pipelines))
	}

	for i := range ipt.CustomerTags {
		customerKeys = append(customerKeys, ipt.CustomerTags[i])
	}
	tags = ipt.Tags
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
