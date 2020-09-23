package neo4j

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "neo4j"

	sampleCfg = `
    # neo4j metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:2004/metrics"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "neo4j"
    ## ignore_measurement = []

    # [inputs.prom.tags]
    # from = "127.0.0.1:2379"
    # tags1 = "value1"
`
)

type Neo4j struct{}

func (*Neo4j) Run() {}

func (*Neo4j) Catalog() string {
	return inputName
}

func (*Neo4j) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Neo4j{}
	})
}
