package druid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "druid"

	defaultMeasurement = "druid"

	sampleCfg = `
[[inputs.druid]]
    # http server route path
    # required
    path = "/druid"

    # [inputs.druid.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Druid{}
	})
}

type Druid struct {
	Path string            `toml:"path"`
	Tags map[string]string `toml:"tags"`
}

func (*Druid) SampleConfig() string {
	return sampleCfg
}

func (*Druid) Catalog() string {
	return inputName
}

func (d *Druid) Run() {
	l = logger.SLogger(inputName)
	l.Infof("druid input started...")
}

func (d *Druid) RegHttpHandler() {
	httpd.RegHttpHandler("POST", d.Path, d.Handle)
}

func (d *Druid) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read body, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := extract(body, d.Tags); err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

type druidMetric []struct {
	Timestamp string  `json:"timestamp"`
	Service   string  `json:"service"`
	Host      string  `json:"host"`
	Version   string  `json:"version"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
}

func extract(body []byte, tags map[string]string) error {
	var metrics druidMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		l.Errorf("failed to paras data, err: %s", err.Error())
		return err
	}

	if len(metrics) == 0 {
		return fmt.Errorf("druid metrics is empty")
	}

	_tags := make(map[string]string)
	_tags["host"] = metrics[0].Host
	_tags["versin"] = metrics[0].Version

	for k, v := range tags {
		_tags[k] = v
	}

	timeNode := getTimeNodeMetrics(metrics)

	flag := false
	for timeKey, fileds := range timeNode {
		t, err := time.Parse(time.RFC3339, timeKey)
		if err != nil {
			l.Errorf("failed to paras timestamp '%s', err: %s", timeKey, err.Error())
			flag = true
			continue
		}

		data, err := io.MakeMetric(defaultMeasurement, _tags, fileds, t)
		if err != nil {
			l.Errorf("failed to make metric, err: %s", err.Error())
			flag = true
			continue
		}

		if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
			l.Errorf("failed to io Feed, err: %s", err.Error())
			flag = true
			continue
		}
	}

	if flag {
		return fmt.Errorf("extract error")
	}

	return nil
}

func getTimeNodeMetrics(metrics druidMetric) map[string]map[string]interface{} {
	var timeNode = make(map[string]map[string]interface{})

	for _, metric := range metrics {

		metricType, ok := metricsTemplate[metric.Metric]
		if !ok {
			continue
		}
		if metric.Service == "druid/peon" {
			// Skipping all metrics from peon. These are task specific and need some
			continue
		}

		timestamp := metric.Timestamp
		if _, ok := timeNode[timestamp]; !ok {
			timeNode[timestamp] = make(map[string]interface{})
		}

		metricKey := strings.ReplaceAll(metric.Service+"."+metric.Metric, "/", ".")
		switch metricType {
		case Normal:
			timeNode[timestamp][metricKey] = metric.Value
		case Count:
			timeNode[timestamp][metricKey] = int64(metric.Value)
		case ConvertRange:
			timeNode[timestamp][metricKey] = metric.Value * 100
		default:
			l.Info("Unknown metric type ", metricType)
		}
	}

	return timeNode
}
