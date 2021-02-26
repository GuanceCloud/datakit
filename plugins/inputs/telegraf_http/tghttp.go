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

// telegraf_http 采集器在不确定使用 pipeline 的正确方式前，不提供 sampleCfg，以便后面全量覆盖

// # [[inputs.telegraf_http.categories]]
// # metric = "A"
// # category = "metric"

)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TelegrafHTTP{
			categoriesMap: make(map[string]string),
		}
	})
}

type TelegrafHTTP struct {
	Categories []struct {
		Measurement string `toml:"metric"`
		Category    string `toml:"category"`
	} `toml:"categories"`

	// map[measurement]io.Metric
	categoriesMap map[string]string
}

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

	for _, c := range t.Categories {
		switch c.Category {
		case "metric":
			t.categoriesMap[c.Measurement] = io.Metric
		case "logging":
			t.categoriesMap[c.Measurement] = io.Logging
		default:
			l.Warnf("invalid category '%s', only accept metric/logging. use default 'metric'", c.Category)
			t.categoriesMap[c.Measurement] = io.Metric
		}
	}

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

		if category, ok := t.categoriesMap[measurement]; !ok {
			// 没有对此 measurement 指定 category，默认发送到 io.Metric
			if err := io.NamedFeed(data, io.Metric, measurement); err != nil {
				l.Error(err)
			}
		} else {
			if err := io.NamedFeed(data, category, measurement); err != nil {
				l.Error(err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
