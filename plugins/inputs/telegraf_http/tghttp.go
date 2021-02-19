package telegraf_http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "telegraf_http"

	sampleCfg = `
[inputs.telegraf_http]

    # [[inputs.telegraf_http.pipeline_metric]]
    # metric = "A"
    # pipeline = "A.p"
    # categories = ["metric", "logging", "object"]
`
)

var (
	l = logger.DefaultSLogger(inputName)
)

type pipelineMetric struct {
	Metric     string   `toml:"metric"`
	Pipeline   string   `toml:"pipeline"`
	Categories []string `toml:"categories,omitempty"`
}

type metric struct {
	pipeline   *pipeline.Pipeline
	categories []string
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TelegrafHTTP{
			pipelineMap: make(map[string]*metric),
		}
	})
}

type TelegrafHTTP struct {
	PipelineMetric []*pipelineMetric `toml:"pipeline_metric"`
	pipelineMap    map[string]*metric
}

func (*TelegrafHTTP) SampleConfig() string {
	return sampleCfg
}

func (*TelegrafHTTP) Catalog() string {
	return inputName
}

func (*TelegrafHTTP) Test() (*inputs.TestResult, error) {
	return &inputs.TestResult{Desc: "success"}, nil
}

// Run() 函数的调用顺序在 RegHttpHandler() 函数之后，
// 可能会出现 RegHttpHandler() 函数执行完毕，HTTP service 已经开启且在接收数据，但 Run() 函数尚未执行，
// 在这种极端情况下：
//     1. 数据会按默认情况处理（即忽略配置文件和 Run() ），
//     2. HTTP goroutine 和执行 Run() 的 goroutine 对同一个 map 做并发读写。
func (t *TelegrafHTTP) Run() {
	l = logger.SLogger(inputName)

	for _, pl := range t.PipelineMetric {
		if _, ok := t.pipelineMap[pl.Metric]; ok {
			l.Warnf("metric '%s' exists", pl.Metric)
			continue
		}

		t.pipelineMap[pl.Metric] = &metric{}

		for _, categroy := range pl.Categories {

			// 从 metric 转换成类似 /v1/write/metric
			c, err := transValidCategory(categroy)
			if err != nil {
				l.Error(err)
				continue
			}

			// 记录有效的 categroy
			t.pipelineMap[pl.Metric].categories = append(t.pipelineMap[pl.Metric].categories, c)
		}

		pipe, err := newPipeline(filepath.Join(datakit.PipelineDir, pl.Pipeline))
		if err != nil {
			l.Error(err) // 忽略error，pipeline对象指针为nil
		} else {
			t.pipelineMap[pl.Metric].pipeline = pipe
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

		m, ok := t.pipelineMap[measurement]
		if !ok {
			// 不对此 measurement 数据做处理，默认发送到 io.Metric
			if err := feed([]byte(point.String()), io.Metric, measurement); err != nil {
				l.Error(err)
			}
			continue
		}

		var data []byte

		if m.pipeline != nil {
			fields, err := point.Fields()
			if err != nil {
				l.Error(err)
				continue
			}

			// 只对行协议的 fields 执行 pipeline，保留原有 tags
			jsonStr, err := fieldsToJSON(fields)
			if err != nil {
				l.Error(err)
				continue
			}

			result, err := m.pipeline.Run(jsonStr).Result()
			if err != nil {
				l.Error(err)
				continue
			}

			pt, err := ifxcli.NewPoint(measurement, point.Tags().Map(), result, point.Time())
			if err != nil {
				l.Error(err)
				continue
			}

			data = []byte(pt.String())

		} else {
			data = []byte(point.String())
		}

		for _, category := range m.categories {
			if err := feed(data, category, measurement); err != nil {
				l.Errorf("feed %s, err: %s", category, err.Error())
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

var categoriesMap = map[string]string{
	"metric":   io.Metric,
	"keyEvent": io.KeyEvent,
	"object":   io.Object,
	"logging":  io.Logging,
	"tracing":  io.Tracing,
	"rum":      io.Rum,
}

func transValidCategory(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("category should not be empty str")
	}

	if c, ok := categoriesMap[s]; ok {
		return c, nil
	}

	return "", fmt.Errorf("invalid category of %s", s)
}

func feed(data []byte, category, name string) error {
	return io.NamedFeed(data, category, name)
}

func newPipeline(pipelinePath string) (*pipeline.Pipeline, error) {
	return pipeline.NewPipelineFromFile(pipelinePath)
}

func fieldsToJSON(fields map[string]interface{}) (string, error) {
	j, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(j), nil
}
