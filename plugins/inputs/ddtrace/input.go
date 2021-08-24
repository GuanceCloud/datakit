package ddtrace

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

// used to set sample ratio
var (
	defRate     = 15
	defScope    = 100
	sampleConfs []*trace.TraceSampleConfig
)

var (
	inputName           = "ddtrace"
	ddtraceSampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## you can stop some patterns by remove them from the list but DO NOT MODIFY THESE PATTERNS.
  # endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with priority 1.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled
    # rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    # [inputs.ddtrace.sample_configs.target]
    # env = "prod"

  ## Sub item in sample_configs list with priority 2.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 100
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 1000
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    # [inputs.ddtrace.sample_configs.target]
    # env = "dev"

    ## ...

  ## Sub item in sample_configs list with priority n.
  # [[inputs.ddtrace.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range that will consider to be covered by sample function.
    # scope = 100
    ## Ignore tags list, tags appear in this list is transparent to sample function that means will always be sampled.
    # ignore_tags_list = []
    ## Sample target, program will search this [tag, value] pair for sampling purpose.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched, so that this pair shoud be empty.
    # [inputs.ddtrace.sample_configs.target]

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
	info, v3, v4, v5, v6 = "/info", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces", "/v0.6/stats"
	defEndpoints         = []string{v3, v4, v5}
	customerTags         []string
	ddTags               map[string]string
	log                  = logger.DefaultSLogger(inputName)
)

type Input struct {
	Path             string                     `toml:"path,omitempty"` // deprecated entry
	Endpoints        []string                   `toml:"endpoints"`
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
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

	sampleConfs = i.TraceSampleConfs

	// validate tracing sample config
	for k, v := range sampleConfs {
		if v.Rate <= 0 || v.Scope < v.Rate {
			v.Rate = 100
			v.Scope = 100
			log.Warnf("%s input tracing sample config [%d] invalid, reset to default.", inputName, k)
		}
	}

	for k := range i.CustomerTags {
		if strings.Contains(i.CustomerTags[k], ".") {
			log.Warn("customer tag can not contains dot(.)")
		} else {
			customerTags = append(customerTags, i.CustomerTags[k])
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
