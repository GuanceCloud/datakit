package telegraf_http

import (
	"io/ioutil"
	"net/http"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "telegraf_http"

	sampleCfg = `
[inputs.telegraf_http]
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TelegrafHTTP{}
	})
}

type TelegrafHTTP struct{}

func (*TelegrafHTTP) Catalog() string {
	return inputName
}

func (*TelegrafHTTP) SampleConfig() string {
	return sampleCfg
}

func (*TelegrafHTTP) Test() (*inputs.TestResult, error) {
	return &inputs.TestResult{Desc: "success"}, nil
}

func (t *TelegrafHTTP) Run() {
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

	for _, point := range points {
		measurement := string(point.Name())
		data := []byte(point.String())

		// 采集器对指定 measurement 的特殊处理
		if fn, ok := globalPointHandle[measurement]; ok {
			if d, err := fn(point); err == nil {
				data = d
			} else {
				l.Warn(err)
				l.Debugf("point handle err, data: %s", point.String())
			}
		}

		if err := io.NamedFeed(data, io.Metric, measurement); err != nil {
			l.Error(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
