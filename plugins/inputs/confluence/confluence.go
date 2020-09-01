package confluence

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "confluence"

	defaultMeasurement = "confluence"

	sampleCfg = `
[[inputs.confluence]]
    # confluence url
    # required
    url = "http://127.0.0.1:8090/plugins/servlet/prometheus/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.confluence.tags]
    # tags1 = "value1"
`
)

var (
	l          = logger.DefaultSLogger(inputName)
	testAssert bool
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Confluence{}
	})
}

type Confluence struct {
	URL      string            `toml:"url"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	duration time.Duration
}

func (*Confluence) SampleConfig() string {
	return sampleCfg
}

func (*Confluence) Catalog() string {
	return inputName
}

func (c *Confluence) Run() {
	l = logger.SLogger(inputName)

	if c.loadcfg() {
		return
	}
	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()

	l.Infof("confluence input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := c.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			if testAssert {
				l.Debugf("data: %s", string(data))
				continue
			}

			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (c *Confluence) loadcfg() bool {
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		c.duration, err = time.ParseDuration(c.Interval)
		if err != nil || c.duration <= 0 {
			l.Errorf("invalid interval, %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}

	return false
}

func (c *Confluence) getMetrics() ([]byte, error) {
	client := &http.Client{}
	client.Timeout = time.Second * 5
	defer client.CloseIdleConnections()

	resp, err := client.Get(c.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.PrometheusToMetrics(resp.Body, inputName, inputName, c.Tags, time.Now())
}
