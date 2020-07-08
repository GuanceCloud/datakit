package coredns

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "coredns"

	defaultMeasurement = "coredns"

	sampleCfg = `
# [[coredns]]
# 	# coredns host
# 	host = "127.0.0.1"
# 	
# 	# coredns prometheus port
# 	port = "9153"
# 	
# 	# second
# 	collect_cycle = 60
# 	
# 	# [inputs.tailf.tags]
# 	# tags1 = "tags1"
`
)

var l *zap.SugaredLogger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Coredns{}
	})
}

type Coredns struct {
	Host    string            `toml:"host"`
	Port    int               `toml:"port"`
	Cycle   time.Duration     `toml:"collect_cycle"`
	Tags    map[string]string `toml:"tags"`
	address string
}

func (_ *Coredns) SampleConfig() string {
	return sampleCfg
}

func (_ *Coredns) Catalog() string {
	return "network"
}

func (c *Coredns) Run() {
	l = logger.SLogger(inputName)

	if _, ok := c.Tags["address"]; !ok {
		c.Tags["address"] = fmt.Sprintf("%s:%d", c.Host, c.Port)
	}

	c.address = fmt.Sprintf("http://%s:%d/metrics", c.Host, c.Port)

	ticker := time.NewTicker(time.Second * c.Cycle)
	defer ticker.Stop()

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
			if err := io.Feed(data, io.Metric); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
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
