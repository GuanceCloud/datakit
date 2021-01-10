package Goruntime

import (
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
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

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &prom.Prom{
			Interval:       datakit.Cfg.MainCfg.Interval,
			InputName:      inputName,
			CatalogStr:     inputName,
			SampleCfg:      sampleCfg,
			Tags:           make(map[string]string),
			IgnoreFunc:     ignore,
			PromToNameFunc: nil,
		}
	})
}

func ignore(pt *ifxcli.Point) bool {
	return false
}
