// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"strings"
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

var (
	inputName    = "ddtrace"
	sampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## customer_tags is a list of keys set by client code like span.SetTag(key, value)
  ## this field will take precedence over [tags] while [customer_tags] merge with [tags].
  ## IT'S EMPTY STRING VALUE AS DEFAULT indicates that no customer tag set up. DO NOT USE DOT(.) IN TAGS
  # customer_tags = []

  ## Keep rare ddtrace resources list.
  # keep_rare_resource = false

  ## Ignore ddtrace resources list. List of strings
  ## A list of regular expressions used to block certain resource name.
  # [inputs.ddtrace.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config
  # [inputs.ddtrace.sampler]
    # priority = 0
    # sampling_rate = 1.0

  ## tags is ddtrace configed key value pairs
  # [inputs.ddtrace.tags]
    # tag1 = "value1"
    # tag2 = "value2"
    # ...
`
	customerKeys []string
	tags         = make(map[string]string)
	log          = logger.DefaultSLogger(inputName)
)

var (
	//nolint: unused,deadcode,varcheck
	info, v3, v4, v5, v6 = "/info", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces", "/v0.6/stats"
	afterGather          = itrace.NewAfterGather()
	keepRareResource     *itrace.KeepRareResource
	closeResource        *itrace.CloseResource
	defSampler           *itrace.Sampler
)

type Input struct {
	Path             string              `toml:"path,omitempty"`           // deprecated
	TraceSampleConfs interface{}         `toml:"sample_configs,omitempty"` // deprecated []*itrace.TraceSampleConfig
	TraceSampleConf  interface{}         `toml:"sample_config"`            // deprecated *itrace.TraceSampleConfig
	IgnoreResources  []string            `toml:"ignore_resources"`         // deprecated []string
	Endpoints        []string            `toml:"endpoints"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	CustomerTags     []string            `toml:"customer_tags"`
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
	dkio.FeedEventLog(&dkio.Reporter{Message: "ddtrace start ok, ready for collecting metrics.", Logtype: "event"})

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

	for k := range ipt.CustomerTags {
		if strings.Contains(ipt.CustomerTags[k], ".") {
			log.Warn("customer tag can not contains dot(.)")
		} else {
			customerKeys = append(customerKeys, ipt.CustomerTags[k])
		}
	}

	if len(ipt.Tags) != 0 {
		tags = ipt.Tags
	}
}

func (ipt *Input) RegHTTPHandler() {
	var isReg bool
	for _, endpoint := range ipt.Endpoints {
		switch endpoint {
		case v3, v4, v5:
			isReg = true
			http.RegHTTPHandler("POST", endpoint, handleTraces(endpoint))
			http.RegHTTPHandler("PUT", endpoint, handleTraces(endpoint))
			log.Infof("pattern %s registered", endpoint)
		case v6:
			isReg = true
			http.RegHTTPHandler("POST", endpoint, handleStats)
			http.RegHTTPHandler("PUT", endpoint, handleStats)
			log.Infof("pattern %s registered", endpoint)
		default:
			log.Errorf("unrecognized ddtrace agent endpoint")
		}
	}
	if isReg {
		itrace.StartTracingStatistic()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
