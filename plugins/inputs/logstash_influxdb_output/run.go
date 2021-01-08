package logstash_influxdb_output

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

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
		"logstash_influxdb_output.p": pipelineSample,
	}
}

func (r *logstashInfluxdbOutput) Run() {
}

func (r *logstashInfluxdbOutput) Test() (result *inputs.TestResult, err error) {
	return
}

func (r *logstashInfluxdbOutput) RegHttpHandler() {
	moduleLogger = logger.SLogger(inputName)
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
		precision = "ns"
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

		pp := process.NewPipeline(r.Pipeline)

		for _, pt := range pts {
			ptname := string(pt.Name())

			m := map[string]interface{}{}
			m["measurement"] = ptname
			m["tags"] = pt.Tags()
			m["time"] = pt.Time().UnixNano()
			fs, _ := pt.Fields()
			for k, v := range fs {
				m[k] = v
			}
			jdata, err := json.Marshal(&m)
			if err != nil {
				moduleLogger.Errorf("%s", err)
				continue
			}
			result := pp.Run(string(jdata)).Result()

			io.NamedFeedEx(inputName, category, ptname, nil, result, pt.Time())
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
