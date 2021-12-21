// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"regexp"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName           = "ddtrace"
	ddtraceSampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## Ignore ddtrace resources list. List of strings
  ## A list of regular expressions filter out certain resource name.
  ## All entries must be double quoted and split by comma.
  # ignore_resources = []

  ## customer_tags is a list of keys set by client code like span.SetTag(key, value)
  ## this field will take precedence over [tags] while [customer_tags] merge with [tags].
  ## IT'S EMPTY STRING VALUE AS DEFAULT indicates that no customer tag set up. DO NOT USE DOT(.) IN TAGS
  # customer_tags = []

  ## tags is ddtrace configed key value pairs
  # [inputs.ddtrace.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    ## ...
`
	customerKeys []string
	ddTags       map[string]string
	log                         = logger.DefaultSLogger(inputName)
	_            inputs.InputV2 = &Input{}
)

var (

	//nolint: unused,deadcode,varcheck
	info, v3, v4, v5, v6 = "/info", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces", "/v0.6/stats"

	ignoreResources []*regexp.Regexp
	filters         []traceFilter
)

type Input struct {
	Path             string                     `toml:"path,omitempty"`           // deprecated
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs,omitempty"` // deprecated
	TraceSampleConf  *trace.TraceSampleConfig   `toml:"sample_config"`            // deprecated
	Endpoints        []string                   `toml:"endpoints"`
	IgnoreResources  []string                   `toml:"ignore_resources"`
	CustomerTags     []string                   `toml:"customer_tags"`
	Tags             map[string]string          `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return ddtraceSampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&DDTraceMeasurement{}}
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	// rare traces penetration
	filters = append(filters, rare)
	// add resource filter
	for k := range i.IgnoreResources {
		if reg, err := regexp.Compile(i.IgnoreResources[k]); err != nil {
			log.Warnf("parse regular expression %q failed", i.IgnoreResources[k])
			continue
		} else {
			ignoreResources = append(ignoreResources, reg)
		}
	}
	if len(ignoreResources) != 0 {
		filters = append(filters, checkResource)
	}
	// add sample filter
	filters = append(filters, sample)

	for k := range i.CustomerTags {
		if strings.Contains(i.CustomerTags[k], ".") {
			log.Warn("customer tag can not contains dot(.)")
		} else {
			customerKeys = append(customerKeys, i.CustomerTags[k])
		}
	}

	if i.Tags != nil {
		ddTags = i.Tags
	} else {
		ddTags = map[string]string{}
	}
}

func (i *Input) RegHTTPHandler() {
	for _, endpoint := range i.Endpoints {
		switch endpoint {
		case v3, v4, v5:
			http.RegHTTPHandler("POST", endpoint, handleTraces(endpoint))
			http.RegHTTPHandler("PUT", endpoint, handleTraces(endpoint))
			log.Infof("pattern %s registered", endpoint)
		case v6:
			http.RegHTTPHandler("POST", endpoint, handleStats)
			http.RegHTTPHandler("PUT", endpoint, handleStats)
			log.Infof("pattern %s registered", endpoint)
		default:
			log.Errorf("unrecognized ddtrace agent endpoint")
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
