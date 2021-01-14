package expressjs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "expressjs"

	sampleCfg = `
[[inputs.prom]]
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

    # [inputs.prom.tags]
    # from = "127.0.0.1:3000"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return prom.NewProm(inputName, inputName, sampleCfg, nil)
	})
}
