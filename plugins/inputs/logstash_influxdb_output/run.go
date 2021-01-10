package logstash_influxdb_output

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	influxm "github.com/influxdata/influxdb1-client/models"
)

var (
	inputName    = `logstash_influxdb_output`
	moduleLogger *logger.Logger
)

func (_ *logstashInfluxdbOutput) Catalog() string {
	return "logstash_influxdb_output"
}

func (_ *logstashInfluxdbOutput) SampleConfig() string {
	return configSample
}

func (r *logstashInfluxdbOutput) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (r *logstashInfluxdbOutput) Run() {
}

func (r *logstashInfluxdbOutput) Test() (result *inputs.TestResult, err error) {
	return
}

func (r *logstashInfluxdbOutput) RegHttpHandler() {
	moduleLogger = logger.SLogger(inputName)

	r.pipelinePool = &sync.Pool{
		New: func() interface{} {
			return pipeline.NewPipeline(r.Pipeline)
		},
	}

	httpd.RegGinHandler("POST", "/write", r.WriteHandler)
	httpd.RegGinHandler("POST", "/ping", PINGHandler)
}

func PINGHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"Code":   200,
		"Status": "Success! Your InfluxDB instance is up and running.",
	})
}

func (r *logstashInfluxdbOutput) WriteHandler(c *gin.Context) {

	body, err := uhttp.GinRead(c)

	if err != nil {
		moduleLogger.Errorf("read http content failed, %s", err)
		uhttp.HttpErr(c, httpd.ErrHttpReadErr)
		return
	}
	defer c.Request.Body.Close()

	if len(body) == 0 {
		moduleLogger.Errorf("empty HTTP body")
		uhttp.HttpErr(c, httpd.ErrBadReq)
		return
	}

	db := c.Query("db")
	precision := c.Query("precision")
	if precision == "" {
		precision = "ms" //https://www.elastic.co/guide/en/logstash/current/plugins-outputs-influxdb.html#plugins-outputs-influxdb-time_precision
	}

	if db == "" {
		moduleLogger.Errorf("db must not be empty")
		uhttp.HttpErr(c, httpd.ErrBadReq)
	}

	moduleLogger.Debugf("influx API query args: %v", c.Request.URL.Query())

	parts := strings.Split(db, ":")

	category := io.Logging
	if len(parts) > 1 {
		switch parts[1] {
		case "metric":
			category = io.Metric
		case "object":
			category = io.Object
		case "event":
			category = io.KeyEvent
		}
	}

	if category == io.Logging {
		pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
		if err != nil {
			moduleLogger.Errorf("fail to parse points, %s", err)
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}

		pp := r.pipelinePool.Get().(*pipeline.Pipeline)
		defer func() {
			r.pipelinePool.Put(pp)
		}()

		for _, pt := range pts {
			ptname := string(pt.Name())

			pipelineInput := map[string]interface{}{}

			rawFields, _ := pt.Fields()
			for k, v := range rawFields {
				pipelineInput[k] = v
			}

			for _, t := range pt.Tags() {
				pipelineInput[string(t.Key)] = string(t.Value)
			}

			pipelineInputBytes, err := json.Marshal(&pipelineInput)
			if err != nil {
				moduleLogger.Warnf("%s", err)
				continue
			}

			moduleLogger.Debugf("pipeline input: %s", string(pipelineInputBytes))
			pipelineResult, err := pp.Run(string(pipelineInputBytes)).Result()
			if err != nil {
				moduleLogger.Errorf("%s", err)
			} else {
				moduleLogger.Debugf("pipeline result: %s", pipelineResult)
				io.NamedFeedEx(inputName, category, ptname, nil, pipelineResult, pt.Time())
			}
		}

	} else {

		if err = io.NamedFeed(body, category, inputName); err != nil {
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &logstashInfluxdbOutput{}
	})
}
