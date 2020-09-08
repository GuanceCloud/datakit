package confluence

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "confluence"

	sampleCfg = `
    # confluence metrics from http://HOST:PORT/plugins/servlet/prometheus/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:8090/plugins/servlet/prometheus/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "confluence"
    ignore_measurement = ["confluence_jvm_info", "confluence_plugin"]

    # [inputs.prom.tags]
    # from = "127.0.0.1:8090"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Confluence{}
	})
}

type Confluence struct {
}

func (*Confluence) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func (*Confluence) Catalog() string {
	return inputName
}

func (*Confluence) Run() {
}
