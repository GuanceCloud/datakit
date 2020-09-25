package etcd

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "etcd"

	sampleCfg = `
    # etcd metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:2379/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    ## Internal configuration. Don't modify.
    name = "etcd"
    ignore_measurement = ["etcd_grpc_server", "etcd_server", "etcd_go_info", "etcd_cluster"]
    
    # [inputs.prom.tags]
    # from = "127.0.0.1:2379"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Etcd{}
	})
}

type Etcd struct {
}

func (*Etcd) SampleConfig() string {
	return "[[inputs.prom]]" + sampleCfg
}

func (*Etcd) Catalog() string {
	return inputName
}

func (*Etcd) Run() {
}
