package telegraf_http

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "telegraf_http"

	sampleCfg = `
[inputs.telegraf_http]

    [inputs.telegraf_http.logging_measurements]
    ## "logging_measurement" = "measurement.p"
`
)

var (
	l = logger.DefaultSLogger(inputName)
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TelegrafHTTP{LoggingMeas: make(map[string]string)}
	})
}

type TelegrafHTTP struct {
	// map[measurement]pipelinePath
	LoggingMeas map[string]string `toml:"logging_measurements"`
}

func (*TelegrafHTTP) SampleConfig() string {
	return sampleCfg
}

func (*TelegrafHTTP) Catalog() string {
	return inputName
}

func (*TelegrafHTTP) Test() (result *inputs.TestResult, err error) {
	result.Desc = "success"
	return
}

func (*TelegrafHTTP) Run() {
	l = logger.SLogger(inputName)
	l.Infof("telegraf_http input started...")
}

func (t *TelegrafHTTP) RegHttpHandler() {
	httpd.RegHttpHandler("POST", "/telegraf", t.Handle)
}

func (t *TelegrafHTTP) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read body, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		l.Debug("empty body")
		return
	}

	points, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), "n")
	if err != nil {
		l.Errorf("parse points, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricFeeds := map[string][]string{}
	loggingFeeds := map[string][]string{}

	for _, point := range points {
		meas := string(point.Name())

		if pipelinePath, ok := t.LoggingMeas[meas]; ok {
			jsonStr, err := inputs.PointToJSON(point)
			if err != nil {
				l.Errorf("point to json, err: %s", err.Error())
				continue
			}

			result := inputs.RunPipeline(jsonStr, pipelinePath)

			pt, err := inputs.MapToPoint(result)
			if err != nil {
				l.Errorf("map to point, err: %s", err.Error())
				continue
			}

			if _, ok := loggingFeeds[meas]; !ok {
				loggingFeeds[meas] = []string{}
			}

			loggingFeeds[meas] = append(loggingFeeds[meas], pt.String())

		} else {
			if _, ok := metricFeeds[meas]; !ok {
				metricFeeds[meas] = []string{}
			}

			metricFeeds[meas] = append(metricFeeds[meas], point.String())
		}
	}

	for k, lines := range metricFeeds {
		if err := io.NamedFeed([]byte(strings.Join(lines, "\n")), io.Metric, k); err != nil {
			l.Errorf("feed metric, err: %s", err.Error())
			return
		}
	}

	for k, lines := range loggingFeeds {
		if err := io.NamedFeed([]byte(strings.Join(lines, "\n")), io.Logging, k); err != nil {
			l.Errorf("feed logging, err: %s", err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
