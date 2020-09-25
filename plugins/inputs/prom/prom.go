package prom

import (
	"bytes"
	"net/http"
	"strings"
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
    name = "temp"

    ## ignore rules
    # ignore_measurement = ["temp_grpc_server""]
    # ignore_tags_key_prefix = []
    # ignore_fields_key_prefix" = []

    # [inputs.prom.tags]
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Prom{}
	})
}

type Prom struct {
	URL                   string            `toml:"url"`
	Interval              string            `toml:"interval"`
	TLSOpen               bool              `toml:"tls_open"`
	CacertFile            string            `toml:"tls_ca"`
	CertFile              string            `toml:"tls_cert"`
	KeyFile               string            `toml:"tls_key"`
	Tags                  map[string]string `toml:"tags"`
	InputName             string            `toml:"name"`
	IgnoreMeasurement     []string          `toml:"ignore_measurement"`
	IgnoreTagsKeyPrefix   []string          `toml:"ignore_tags_key_prefix"`
	IgnoreFieldsKeyPrefix []string          `toml:"ignore_fields_key_prefix"`

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
	if p.loadcfg() {
		return
	}

	ticker := time.NewTicker(p.duration)
	defer ticker.Stop()
	p.log.Infof("%s input started.", p.InputName)

	for {
		select {
		case <-datakit.Exit.Wait():
			p.client.CloseIdleConnections()
			p.log.Info("exit")
			return

		case <-ticker.C:
			data, err := p.getMetrics()
			if err != nil {
				p.log.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, p.InputName); err != nil {
				p.log.Error(err)
				continue
			}
			p.log.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (p *Prom) loadcfg() bool {
	var err error
	p.log = logger.SLogger(p.InputName)
	p.client = &http.Client{}
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

	pts, err := cliutils.PromTextToMetrics(resp.Body, p.InputName, p.InputName, time.Now())
	if err != nil {
		return nil, err
	}
	var buffer = bytes.Buffer{}

	for _, pt := range pts {
		if p.ignore(pt) {
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

func (p *Prom) ignore(pt *ifxcli.Point) bool {
	fields, err := pt.Fields()
	if err != nil {
		p.log.Error(err)
		return true
	}
	return p.ignoreMeasurement(pt.Name()) || p.ignoreTagsKeyPrefix(pt.Tags()) || p.ignoreFieldsKeyPrefix(fields)
}

func (p *Prom) ignoreMeasurement(measurement string) bool {
	for _, m := range p.IgnoreMeasurement {
		if measurement == m {
			return true
		}
	}
	return false
}

func (p *Prom) ignoreTagsKeyPrefix(tags map[string]string) bool {
	if len(p.IgnoreTagsKeyPrefix) == 0 {
		return false
	}
	for key := range tags {
		for _, m := range p.IgnoreTagsKeyPrefix {
			if strings.HasPrefix(key, m) {
				return true
			}
		}
	}
	return false
}

func (p *Prom) ignoreFieldsKeyPrefix(fields map[string]interface{}) bool {
	if len(p.IgnoreFieldsKeyPrefix) == 0 {
		return false
	}
	for key := range fields {
		for _, m := range p.IgnoreFieldsKeyPrefix {
			if strings.HasPrefix(key, m) {
				return true
			}
		}
	}
	return false
}
