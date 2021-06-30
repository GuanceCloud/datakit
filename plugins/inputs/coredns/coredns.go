package coredns

import (
	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "coredns"

	sampleCfg = `
[[inputs.prom]]
    ## coredns metrics from http://HOST:PORT/metrics
    ## usually modify host and port
    url = "http://127.0.0.1:9153/metrics"
    
    ## valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    interval = "10s"
    
    [inputs.prom.tags]
    # from = "127.0.0.1:9153"
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return prom.NewProm(inputName, "network", sampleCfg, ignore)
	})
}

var ignoreMeasurements = []string{
	"coredns_plugin",
	"coredns_build",
	"coredns_go_info",
}

func ignore(pt *ifxcli.Point) bool {
	for _, m := range ignoreMeasurements {
		if pt.Name() == m {
			return true
		}
	}
	return false
}
