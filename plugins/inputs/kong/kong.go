package kong

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "kong"

	sampleCfg = `
    # neo4j metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:8001/metrics"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "kong"
    ## ignore_measurement = []

    # [inputs.prom.tags]
    # from = "127.0.0.1:2379"
    # tags1 = "value1"
`
)

type Kong struct{}

func (*Kong) Run() {}

func (*Kong) Catalog() string {
	return inputName
}

func (*Kong) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Kong{}
	})
}
