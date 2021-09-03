package ddtrace

import (
	"regexp"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	inputName           = "ddtrace"
	ddtraceSampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## you can stop some patterns by remove them from the list but DO NOT MODIFY THESE PATTERNS.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## Tracing data sample config, [rate] and [scope] together determine how many trace data
  ## will be send to DataFlux workspace.
  # [inputs.ddtrace.sample_config]
    ## Sample rate, how many tracing data will be sampled
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []

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
	customerKeys  []string
	ddTags        map[string]string
	defIgnoreTags = map[string]string{"_dd.origin": "rum"}
	log           = logger.DefaultSLogger(inputName)
)

var (
	info, v3, v4, v5, v6 = "/info", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces", "/v0.6/stats"
	defEndpoints         = []string{v3, v4, v5}
	ignoreResources      []*regexp.Regexp
	defRate              = 15
	defScope             = 100
	sampleConf           *trace.TraceSampleConfig
	filters              []traceFilter
)

type Input struct {
	Path             string                     `toml:"path,omitempty"` // deprecated entry
	Endpoints        []string                   `toml:"endpoints"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs,omitempty"` // deprecated in new issue
	TraceSampleConf  *trace.TraceSampleConfig   `toml:"sample_config"`
	IgnoreResources  []string                   `toml:"ignore_resources"`
	CustomerTags     []string                   `toml:"customer_tags"`
	Tags             map[string]string          `toml:"tags"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return ddtraceSampleConfig
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&DdtraceMeasurement{}}
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	for k := range i.IgnoreResources {
		if reg, err := regexp.Compile(i.IgnoreResources[k]); err != nil {
			log.Warnf("parse regular expression %q failed", i.IgnoreResources[k])
			continue
		} else {
			ignoreResources = append(ignoreResources, reg)
		}
	}
	if len(ignoreResources) != 0 {
		filters = append(filters, filterOutResource)
	}

	sampleConf = i.TraceSampleConf
	if sampleConf != nil {
		if sampleConf.Rate <= 0 || sampleConf.Scope < sampleConf.Rate {
			sampleConf.Rate = 100
			sampleConf.Scope = 100
			log.Warnf("input tracing sample config invalid, fallback to fully sampling.")
		}
		filters = append(filters, sample)
	}

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

func (i *Input) RegHttpHandler() {
	if len(i.Endpoints) == 0 {
		i.Endpoints = defEndpoints
	}

	// // do not register /info in endpoints
	// http.RegHttpHandler("GET", info, handleInfo)

	for _, endpoint := range i.Endpoints {
		switch endpoint {
		case v3, v4, v5:
			http.RegHttpHandler("POST", endpoint, handleTraces(endpoint))
			http.RegHttpHandler("PUT", endpoint, handleTraces(endpoint))
			log.Infof("pattern %s registered")
		case v6:
			http.RegHttpHandler("POST", endpoint, handleStats)
			http.RegHttpHandler("PUT", endpoint, handleStats)
			log.Infof("pattern %s registered")
		default:
			log.Errorf("unrecognized ddtrace agent endpoint")
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
