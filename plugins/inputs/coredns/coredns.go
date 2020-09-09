package coredns

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "coredns"

	sampleCfg = `
    # coredns metrics from http://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:9153/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # Internal configurationl. Don't modify
    name = "coredns"
    ignore_measurement = ["coredns_plugin", "coredns_build", "coredns_go_info"]

    # [inputs.prom.tags]
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
}

func (*Coredns) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func (*Coredns) Catalog() string {
	return "network"
}

func (*Coredns) Run() {
}
