package envoy

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "envoy"

	defaultMeasurement = "envoy"

	sampleCfg = `
[[inputs.envoy]]
    # envoy metrics from http(https)://HOST:PORT/stats/prometheus
    # usually modify host and port
    # required
    url = "http://127.0.0.1:8090/stats/prometheus"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"
    
    # [inputs.envoy.tags]
    # from = "127.0.0.1:9901"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Envoy{}
	})
}

type Envoy struct {
	URL        string            `toml:"url"`
	Interval   string            `toml:"interval"`
	CacertFile string            `toml:"tls_ca"`
	CertFile   string            `toml:"tls_cert"`
	KeyFile    string            `toml:"tls_key"`
	Tags       map[string]string `toml:"tags"`
	TLSOpen    bool              `toml:"tls_open"`
}

func (*Envoy) SampleConfig() string {
	return sampleCfg
}

func (*Envoy) Catalog() string {
	return inputName
}

func (e *Envoy) Run() {
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

		IgnoreMeasurement:     []string{"envoy_http", "envoy_listener"},
		IgnoreFieldsKeyPrefix: []string{"envoy_server_worker"},
	}
	p.Start()
}
