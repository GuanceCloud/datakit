package rum

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
)

const (
	PRECISION         = "precision"
	DEFAULT_PRECISION = "ns"
)

var (
	inputName                   = `rum`
	ipheaderName                = ""
	l            *logger.Logger = logger.DefaultSLogger(inputName)
)

func (_ *Rum) Catalog() string {
	return "rum"
}

func (_ *Rum) SampleConfig() string {
	return configSample
}

func (r *Rum) Run() {
}

func (r *Rum) Test() (result *inputs.TestResult, err error) {
	return
}

func (r *Rum) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (r *Rum) RegHttpHandler() {
	ipheaderName = r.IPHeader
	l = logger.SLogger(inputName)
	httpd.RegGinHandler("POST", io.Rum, r.Handle)
}

func (r *Rum) Handle(c *gin.Context) {

	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error
	sourceIP := ""

	precision, _ = uhttp.GinGetArg(c, "X-Precision", PRECISION)

	if ipheaderName != "" {
		sourceIP = c.Request.Header.Get(ipheaderName)
		if sourceIP != "" {
			parts := strings.Split(sourceIP, ",")
			if len(parts) > 0 {
				sourceIP = parts[0]
			}
		}
	}

	if sourceIP == "" {
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 0 {
			sourceIP = parts[0]
		}
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrHttpReadErr, err.Error()))
		return
	}

	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received %d points", len(pts))

	metricsdata := [][]byte{}
	esdata := [][]byte{}

	pp := pipeline.NewPipeline(r.Pipeline)

	for _, pt := range pts {
		ptname := string(pt.Name())

		m := map[string]string{
			"ip": sourceIP,
		}

		jdata, _ := json.Marshal(&m)
		l.Debugf("input data: %s", string(jdata))

		result, err := pp.Run(string(jdata)).Result()
		if err != nil {
			l.Warnf("pipeline run error, %s", err)
		} else {
			l.Debugf("pipeline result: %s", result)
		}

		for k, v := range result {
			if k != "ip" {
				if sv, ok := v.(string); ok {
					pt.AddTag(k, sv)
				}
			}
		}

		line := []byte(pt.String())

		if IsMetric(ptname) {
			metricsdata = append(metricsdata, line)
		} else if IsES(ptname) {
			esdata = append(esdata, line)
		} else {
			uhttp.HttpErr(c, uhttp.Errorf(httpd.ErrBadReq, "unknown RUM metric name `%s'", ptname))
			return
		}
	}

	if len(metricsdata) > 0 {
		body = bytes.Join(metricsdata, []byte("\n"))
		if err = io.NamedFeed(body, io.Metric, inputName); err != nil {
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}
	}

	if len(esdata) > 0 {
		body = bytes.Join(esdata, []byte("\n"))

		if err = io.NamedFeed(body, io.Rum, inputName); err != nil {
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}
	}

	httpd.ErrOK.HttpBody(c, nil)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Rum{}
	})
}
