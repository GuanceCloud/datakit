package confluence

import (
	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "confluence"

	sampleCfg = `
[[inputs.prom]]
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

    # [inputs.prom.tags]
    # from = "127.0.0.1:8090"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return prom.NewProm(inputName, inputName, sampleCfg, ignore)
	})
}

var ignoreMeasurements = []string{
	"confluence_jvm_info",
	"confluence_plugin",
}

func ignore(pt *ifxcli.Point) bool {
	for _, m := range ignoreMeasurements {
		if pt.Name() == m {
			return true
		}
	}
	return false
}
