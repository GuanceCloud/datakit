package expressjs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "expressjs"

	sampleCfg = `
    # expressjs metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:3000/metrics"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "expressjs"
    ## ignore_measurement = ['nodejs_version_info']

    # [inputs.prom.tags]
    # from = "127.0.0.1:2379"
    # tags1 = "value1"
`
)

type ExpressJS struct{}

func (e *ExpressJS) Run() {}

func (*ExpressJS) Catalog() string {
	return inputName
}

func (*ExpressJS) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &ExpressJS{}
	})
}
