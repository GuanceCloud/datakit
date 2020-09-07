package etcd

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "etcd"

	defaultMeasurement = "etcd"

	sampleCfg = `
[[inputs.etcd]]
    # etcd metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:2379/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/path/to/ca.crt"
    # tls_cert = "/path/to/peer.crt"
    # tls_key = "/path/to/peer.key"
    
    # [inputs.etcd.tags]
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
	URL        string            `toml:"url"`
	Interval   string            `toml:"interval"`
	CacertFile string            `toml:"tls_cacert_file"`
	CertFile   string            `toml:"tls_cert_file"`
	KeyFile    string            `toml:"tls_key_file"`
	Tags       map[string]string `toml:"tags"`
	TLSOpen    bool              `toml:"tls_open"`
}

func (*Etcd) SampleConfig() string {
	return sampleCfg
}

func (*Etcd) Catalog() string {
	return inputName
}

func (e *Etcd) Run() {
	p := prom.Prom{
		URL:        e.URL,
		Interval:   e.Interval,
		TLSOpen:    e.TLSOpen,
		CacertFile: e.CacertFile,
		CertFile:   e.CertFile,
		KeyFile:    e.KeyFile,
		Tags:       e.Tags,

		InputName:          inputName,
		DefaultMeasurement: defaultMeasurement,

		IgnoreMeasurement: []string{"etcd_grpc_server", "etcd_server", "etcd_go_info", "etcd_cluster"},
	}
	p.Start()
}
