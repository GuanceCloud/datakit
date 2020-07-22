package druid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// druid config
// bin: datakit_test/druid/apache-druid-0.18.1/bin/start-micro-quickstart
// config file: druid/apache-druid-0.18.1/conf/druid/single-server/micro-quickstart/_common/common.runtime.properties
//
// druid.monitoring.emissionPeriod=PT10s
// druid.monitoring.monitors=["com.metamx.metrics.JvmMonitor"]
// druid.emitter=none
// druid.emitter.http.flushMillis=10000
// druid.emitter.http.recipientBaseUrl=http://ADDR_TO_THIS_SERVICE:8424

const (
	inputName = "druid"

	defaultMeasurement = "druid"

	sampleCfg = `
# [inputs.druid]
# 	# [inputs.druid.tags]
# 	# tags1 = "tags1"
`
)

var (
	l *logger.Logger

	testAssert bool

	globalTags map[string]string
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Druid{}
	})
}

type Druid struct {
	Tags map[string]string `toml:"tags"`
}

func (d *Druid) SampleConfig() string {
	return sampleCfg
}

func (d *Druid) Catalog() string {
	return inputName
}

func (d *Druid) Run() {
	l = logger.SLogger(inputName)
	l.Infof("druid input started...")

	if globalTags == nil {
		globalTags = d.Tags
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read body, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := extract(body); err == nil {
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

func extract(body []byte) error {

	var metrics druidMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		l.Errorf("failed to paras data, err: %s", err.Error())
		return err
	}

	var timeNode = make(map[string]map[string]interface{})
	var host, version string
	if len(metrics) > 0 {
		host = metrics[0].Host
		version = metrics[0].Version
	} else {
		return fmt.Errorf("druid metrics is empty")
	}

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

		metricKey := strings.Replace(metric.Service+"."+metric.Metric, "/", ".", -1)
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

	tags := make(map[string]string)
	tags["host"] = host
	tags["versin"] = version
	for k, v := range globalTags {
		tags[k] = v
	}

	flag := false
	for timeKey, fileds := range timeNode {
		t, err := time.Parse(time.RFC3339, timeKey)
		if err != nil {
			l.Errorf("failed to paras timestamp '%s', err: %s", timeKey, err.Error())
			flag = true
			continue
		}

		data, err := io.MakeMetric(defaultMeasurement, tags, fileds, t)
		if err != nil {
			l.Errorf("failed to make metric, err: %s", err.Error())
			flag = true
			continue
		}

		if testAssert {
			fmt.Println(string(data))
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
