package confluence

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "confluence"

	defaultMeasurement = "confluence"

	sampleCfg = `
[[inputs.confluence]]
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
    
    # [inputs.confluence.tags]
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
	URL        string            `toml:"url"`
	Interval   string            `toml:"interval"`
	CacertFile string            `toml:"tls_ca"`
	CertFile   string            `toml:"tls_cert"`
	KeyFile    string            `toml:"tls_key"`
	Tags       map[string]string `toml:"tags"`
	TLSOpen    bool              `toml:"tls_open"`
}

func (*Confluence) SampleConfig() string {
	return sampleCfg
}

func (*Confluence) Catalog() string {
	return inputName
}

func (c *Confluence) Run() {
	p := prom.Prom{
		URL:        c.URL,
		Interval:   c.Interval,
		TLSOpen:    c.TLSOpen,
		CacertFile: c.CacertFile,
		CertFile:   c.CertFile,
		KeyFile:    c.KeyFile,
		Tags:       c.Tags,

		InputName:          inputName,
		DefaultMeasurement: defaultMeasurement,

		IgnoreMeasurement: []string{"confluence_jvm_info", "confluence_plugin"},
	}
	p.Start()
}
