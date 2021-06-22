package logstash_influxdb_output

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

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

// TODO
func (*logstashInfluxdbOutput) RunPipeline() {
}

func (r *logstashInfluxdbOutput) Run() {
}

func (r *logstashInfluxdbOutput) RegHttpHandler() {
	moduleLogger = logger.SLogger(inputName)

	script := r.Pipeline
	if script == "" {
		scriptPath := filepath.Join(datakit.PipelineDir, inputName+".p")
		data, err := ioutil.ReadFile(scriptPath)
		if err == nil {
			script = string(data)
		}
	}

	r.pipelinePool = &sync.Pool{
		New: func() interface{} {
			if script == "" {
				return nil
			}
			p, err := pipeline.NewPipeline(script)
			if err != nil {
				moduleLogger.Errorf("%s", err)
				return nil
			}
			return p
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

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			moduleLogger.Errorf("panic: %s", err)
			moduleLogger.Errorf("%s", string(buf[:n]))
		}
	}()

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

	category := datakit.Logging
	if len(parts) > 1 {
		switch parts[1] {
		case "metric":
			category = datakit.Metric
		case "object":
			category = datakit.Object
		case "event":
			category = io.KeyEvent
		}
	}

	if category == datakit.Logging {
		pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
		if err != nil {
			moduleLogger.Errorf("fail to parse points, %s", err)
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}

		pp_ := r.pipelinePool.Get()
		var pp *pipeline.Pipeline
		if pp_ != nil {
			pp = pp_.(*pipeline.Pipeline)
		}
		defer func() {
			if pp != nil {
				r.pipelinePool.Put(pp)
			}
		}()

		for _, pt := range pts {
			ptname := string(pt.Name())

			if pp == nil {
				io.NamedFeed([]byte(pt.String()), category, inputName)
				continue
			}

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
				moduleLogger.Errorf("%s", err)
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
