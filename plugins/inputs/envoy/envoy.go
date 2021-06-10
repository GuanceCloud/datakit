package envoy

import (
	"strings"

	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "envoy"

	sampleCfg = `
[[inputs.prom]]
    ## envoy metrics from http(https)://HOST:PORT/stats/prometheus
    ## usually modify host and port
    ## required
    url = "http://127.0.0.1:8090/stats/prometheus"

    ## valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    ## required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    [inputs.prom.tags]
    # from = "127.0.0.1:9901"
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return prom.NewProm(inputName, inputName, sampleCfg, ignore)
	})
}

var (
	ignoreMeasurements    = []string{"envoy_http", "envoy_listener"}
	ignoreFieldsKeyPrefix = []string{"envoy_server_worker"}
)

func ignore(pt *ifxcli.Point) bool {
	fields, err := pt.Fields()
	if err != nil {
		return false
	}

	for _, m := range ignoreMeasurements {
		if pt.Name() == m {
			return true
		}
	}

	for key := range fields {
		for _, m := range ignoreFieldsKeyPrefix {
			if strings.HasPrefix(key, m) {
				return true
			}
		}
	}
	return false
}
