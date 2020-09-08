package coredns

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "coredns"

	defaultMeasurement = "coredns"

	sampleCfg = `
[[inputs.coredns]]
    # coredns metrics from http://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:9153/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.coredns.tags]
    # from = "127.0.0.1:9153"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Coredns{}
	})
}

type Coredns struct {
	URL      string            `toml:"url"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`
}

func (*Coredns) SampleConfig() string {
	return sampleCfg
}

func (*Coredns) Catalog() string {
	return "network"
}

func (c *Coredns) Run() {
	p := prom.Prom{
		URL:      c.URL,
		Interval: c.Interval,
		Tags:     c.Tags,

		InputName:          inputName,
		DefaultMeasurement: defaultMeasurement,

		IgnoreMeasurement: []string{"coredns_plugin", "coredns_build", "coredns_go_info"},
	}
	p.Start()
}
