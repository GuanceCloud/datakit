package coredns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "coredns"

	configSample = `
# [[coredns]]
#       ## coredns 地址
#	host = "127.0.0.1"
#
#       ## coredns prometheus 监控端口
#	port = "9153"
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
#
#       ## measurement，不可重复
#       measurement = "coredns"
`
)

var l *zap.SugaredLogger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Coredns{}
	})
}

type (
	Coredns struct {
		C []Impl `toml:"coredns"`
	}

	Impl struct {
		Host        string        `toml:"host"`
		Port        int           `toml:"port"`
		Cycle       time.Duration `toml:"collect_cycle"`
		Measurement string        `toml:"measurement"`
		address     string
	}
)

func (_ *Coredns) SampleConfig() string {
	return configSample
}

func (_ *Coredns) Catalog() string {
	return "network"
}

func (c *Coredns) Run() {
	l = logger.SLogger(inputName)

	for _, i := range c.C {
		go i.start()
	}
}

var tagsWhiteList = map[string]byte{"version": '0'}

func (i *Impl) start() {
	i.address = fmt.Sprintf("http://%s:%d/metrics", i.Host, i.Port)

	if i.Measurement == "" {
		l.Error("invalid measurement")
		return
	}

	ticker := time.NewTicker(time.Second * i.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			pt, err := i.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}

			io.Feed([]byte(pt.String()), io.Metric)
		}
	}
}

func (i *Impl) getMetrics() (*influxdb.Point, error) {

	resp, err := http.Get(i.address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, errors.New("the metrics is empty")
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {

		for k, v := range metric.Tags() {
			if _, ok := tagsWhiteList[k]; ok {
				tags[k] = v
			}
		}

		for k, v := range metric.Fields() {
			fields[k] = v
		}

	}

	return influxdb.NewPoint(i.Measurement, tags, fields, metrics[0].Time())
}
