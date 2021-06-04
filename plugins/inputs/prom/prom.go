package prom

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "prom"

	sampleCfg = `
[[inputs.prom]]
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

    ## data source
    # required
    source = ""

    [inputs.prom.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var (
	defaultIgnoreFunc     = func(*ifxcli.Point) bool { return false }
	defaultPromToNameFunc = func(old string) (string, string, error) { return old, old, nil }
)

type Prom struct {
	URL      string `toml:"url"`
	Interval string `toml:"interval"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	Tags map[string]string `toml:"tags"`

	InputName  string `toml:"source"`
	CatalogStr string
	SampleCfg  string

	IgnoreFunc     func(*ifxcli.Point) bool
	PromToNameFunc func(old string) (string, string, error)

	client   *http.Client
	duration time.Duration
	log      *logger.Logger
}

func NewProm(inputName, catalogStr, sampleCfg string, ignoreFunc func(*ifxcli.Point) bool) *Prom {
	return &Prom{
		InputName:      inputName,
		CatalogStr:     inputName,
		SampleCfg:      sampleCfg,
		Interval:       datakit.Cfg.IntervalDeprecated,
		Tags:           make(map[string]string),
		IgnoreFunc:     ignoreFunc,
		PromToNameFunc: nil,
	}
}

func (p *Prom) SampleConfig() string {
	return p.SampleCfg
}

func (p *Prom) Catalog() string {
	if p.CatalogStr == "" {
		return "prom"
	}

	return p.CatalogStr
}

func (p *Prom) Run() {
	p.log = logger.SLogger(p.InputName)

	if p.initCfg() {
		return
	}
	defer p.stop()

	ticker := time.NewTicker(p.duration)
	defer ticker.Stop()

	p.log.Infof("%s input started.", p.InputName)

	for {
		select {
		case <-datakit.Exit.Wait():
			p.log.Info("exit")
			return

		case <-ticker.C:
			data, err := p.getMetrics()
			if err != nil {
				p.log.Error(err)
				continue
			}
			if err := io.NamedFeed(data, datakit.Metric, p.InputName); err != nil {
				p.log.Error(err)
				continue
			}
			p.log.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (p *Prom) stop() {
	p.client.CloseIdleConnections()
}

func (p *Prom) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			p.log.Info("exit")
			return true
		default:
			// nil
		}

		if err := p.loadCfg(); err != nil {
			p.log.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return false
}

func (p *Prom) loadCfg() (err error) {
	p.duration, err = time.ParseDuration(p.Interval)
	if err != nil {
		err = fmt.Errorf("invalid interval, %s", err.Error())
		return
	} else if p.duration <= 0 {
		err = fmt.Errorf("invalid interval, cannot be less than zero")
		return
	}

	p.client = &http.Client{
		Timeout: 5 * time.Second,
	}
	if p.TLSOpen {
		tc, _err := TLSConfig(p.CacertFile, p.CertFile, p.KeyFile)
		if _err != nil {
			return _err
		} else {
			p.client.Transport = &http.Transport{
				TLSClientConfig: tc,
			}
		}
	}

	return
}

func (p *Prom) getMetrics() ([]byte, error) {
	resp, err := p.client.Get(p.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pts, err := cliutils.PromTextToMetrics(resp.Body, p.InputName, p.InputName, time.Now())
	if err != nil {
		return nil, err
	}
	var buffer = bytes.Buffer{}

	for _, pt := range pts {
		if p.IgnoreFunc != nil && p.IgnoreFunc(pt) {
			continue
		}

		fields, _ := pt.Fields()
		tags := pt.Tags()
		for k, v := range p.Tags {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}
		data, err := io.MakeMetric(pt.Name(), tags, fields, pt.Time())
		if err != nil {
			continue
		}

		buffer.Write(data)
		buffer.WriteString("\n")
	}

	return buffer.Bytes(), nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Prom{
			Interval:       datakit.Cfg.IntervalDeprecated,
			InputName:      inputName,
			SampleCfg:      sampleCfg,
			Tags:           make(map[string]string),
			CatalogStr:     "prom",
			IgnoreFunc:     defaultIgnoreFunc,
			PromToNameFunc: defaultPromToNameFunc,
		}
	})
}
