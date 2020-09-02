package prom

import (
	"bytes"
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

	defaultMeasurement = "prom"

	sampleCfg = `
[[inputs.prom]]
    # required
    url = "127.0.0.1:9090/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # use HTTPS TLS
    tls_open = false
    tls_ca = "ca.crt"
    tls_cert = "peer.crt"
    tls_key = "peer.key"
    
    # [inputs.prom.tags]
    # tags1 = "value1"
`
)

// var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Prom{}
	})
}

type Prom struct {
	URL        string            `toml:"url"`
	Interval   string            `toml:"interval"`
	TLSOpen    bool              `toml:"tls_open"`
	CacertFile string            `toml:"tls_ca"`
	CertFile   string            `toml:"tls_cert"`
	KeyFile    string            `toml:"tls_key"`
	Tags       map[string]string `toml:"tags"`

	InputName          string
	DefaultMeasurement string

	IgnoreMeasurement []string

	client   *http.Client
	duration time.Duration
	log      *logger.Logger
}

func (*Prom) SampleConfig() string {
	return sampleCfg
}

func (*Prom) Catalog() string {
	return inputName
}

func (p *Prom) Run() {
	p.InputName = inputName
	p.DefaultMeasurement = defaultMeasurement
	p.Start()
}

func (p *Prom) Start() {
	p.log = logger.SLogger(p.InputName)
	p.client = &http.Client{}
	defer p.client.CloseIdleConnections()

	if p.loadcfg() {
		return
	}

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
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				p.log.Error(err)
				continue
			}
			p.log.Debugf("feed %d bytes to io ok", len(data))
		}
	}

}

func (p *Prom) loadcfg() bool {
	var err error
	p.client.Timeout = time.Second * 5

	for {
		select {
		case <-datakit.Exit.Wait():
			p.log.Info("exit")
			return true
		default:
			// nil
		}

		p.duration, err = time.ParseDuration(p.Interval)
		if err != nil || p.duration <= 0 {
			p.log.Errorf("invalid interval, err %s", err.Error())
			time.Sleep(time.Second)
			continue
		}

		if p.TLSOpen {
			tc, err := TLSConfig(p.CacertFile, p.CertFile, p.KeyFile)
			if err != nil {
				p.log.Error(err)
				time.Sleep(time.Second)
			} else {
				p.client.Transport = &http.Transport{
					TLSClientConfig: tc,
				}
				break
			}
		} else {
			break
		}
	}
	return false
}

func (p *Prom) getMetrics() ([]byte, error) {
	resp, err := p.client.Get(p.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pts, err := cliutils.PromTextToMetrics(resp.Body, p.InputName, p.DefaultMeasurement, time.Now())
	if err != nil {
		return nil, err
	}
	var buffer = bytes.Buffer{}

	for _, pt := range pts {
		if p.ignore(pt) {
			continue
		}

		fields, err := pt.Fields()
		if err != nil {
			p.log.Error(err)
			continue
		}

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

func (p *Prom) ignore(pt *ifxcli.Point) bool {
	return p.ignoreMeasurement(pt.Name())
}

func (p *Prom) ignoreMeasurement(measurement string) bool {
	for _, m := range p.IgnoreMeasurement {
		if measurement == m {
			return true
		}
	}
	return false
}
