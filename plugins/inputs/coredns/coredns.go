package coredns

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "coredns"

	defaultMeasurement = "coredns"

	sampleCfg = `
# [[inputs.coredns]]
# 	# coredns host
#	# required
# 	host = "127.0.0.1"
#
# 	# coredns prometheus port
#	# required
# 	port = 9153
#
# 	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
#	# required
# 	interval = "10s"
#
# 	# [inputs.coredns.tags]
# 	# tags1 = "value1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Coredns{}
	})
}

type Coredns struct {
	Host     string            `toml:"host"`
	Port     int               `toml:"port"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`
	address  string

	// forward compatibility
	CollectCycle string `toml:"collect_cycle"`

	duration time.Duration
}

func (_ *Coredns) SampleConfig() string {
	return sampleCfg
}

func (_ *Coredns) Catalog() string {
	return "network"
}

func (c *Coredns) Run() {
	l = logger.SLogger(inputName)

	if c.loadcfg() {
		return
	}
	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()

	l.Infof("coredns input started.")

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
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (c *Coredns) loadcfg() bool {

	if c.Interval == "" && c.CollectCycle != "" {
		c.Interval = c.CollectCycle
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d, err := time.ParseDuration(c.CollectCycle)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval, %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		c.duration = d
		break
	}

	if c.Tags == nil {
		c.Tags = make(map[string]string)
	}

	if _, ok := c.Tags["address"]; !ok {
		c.Tags["address"] = fmt.Sprintf("%s:%d", c.Host, c.Port)
	}

	c.address = fmt.Sprintf("http://%s:%d/metrics", c.Host, c.Port)

	return false
}

func (c *Coredns) getMetrics() ([]byte, error) {

	resp, err := http.Get(c.address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("the metrics is empty")
	}

	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {
		for k, v := range metric.Fields() {
			fields[k] = v
		}
	}

	return io.MakeMetric(defaultMeasurement, c.Tags, fields, time.Now())
}
