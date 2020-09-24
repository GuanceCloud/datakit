package Goruntime

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "goruntime"

	sampleCfg = `
    # Goruntime metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:9090/metrics"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "goruntime"
    ## ignore_measurement = []

    # [inputs.prom.tags]
    # from = "127.0.0.1:9090"
    # tags1 = "value1"
`
)

type Goruntime struct{}

func (*Goruntime) Run() {}

func (*Goruntime) Catalog() string {
	return "golang"
}

func (*Goruntime) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Goruntime{}
	})
}
