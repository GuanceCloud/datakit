package kong

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l    *logger.Logger
	name = "kong"
)

func (_ *Kong) Catalog() string {
	return "kong"
}

func (_ *Kong) SampleConfig() string {
	return configSample
}

func (_ *Kong) Description() string {
	return ""
}

func (_ *Kong) Gather() error {
	return nil
}

func (_ *Kong) Init() error {
	return nil
}

func (k *Kong) Run() {
	l = logger.SLogger("baiduIndex")

	l.Info("baiduIndex input started...")

	if k.initcfg() {
		return
	}

	interval, err := time.ParseDuration(k.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			k.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (k *Kong) initcfg() bool {
	if k.Tags == nil {
		k.Tags = make(map[string]string)
	}

	if k.MetricName == "" {
		k.MetricName = "kong"
	}

	return false
}

func (kong *Kong) handle() {
	client := &http.Client{}
	client.Timeout = time.Second * 5
	defer client.CloseIdleConnections()

	resp, err := client.Get(kong.Addr)
	if err != nil {
		l.Errorf("get metric from kong error %s", err)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		l.Errorf("prom metric convert influxdb point error %s", err)
	}

	if len(metrics) == 0 {
		l.Error("metrics is empty")
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {
		for k, v := range metric.Tags() {
			tags[k] = v
		}

		for k, v := range metric.Fields() {
			fields[k] = v
		}
	}

	for k, v := range kong.Tags {
		tags[k] = v
	}

	pt, err := io.MakeMetric(kong.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %s", err)
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %s", err)
	}
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &Kong{}
	})
}
