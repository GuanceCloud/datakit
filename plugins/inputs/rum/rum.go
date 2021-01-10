package rum

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
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

	r.pipelinePool = &sync.Pool{
		New: func() interface{} {
			return pipeline.NewPipeline(r.Pipeline)
		},
	}

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

	pp := r.pipelinePool.Get().(*pipeline.Pipeline)
	defer func() {
		r.pipelinePool.Put(pp)
	}()

	for _, pt := range pts {
		ptname := string(pt.Name())

		if IsMetric(ptname) {

			ipInfo, err := geo.Geo(sourceIP)
			if err != nil {
				l.Errorf("parse ip error: %s", err)
			} else {
				pt.AddTag("city", ipInfo.City)
				pt.AddTag("region", ipInfo.Region)
				pt.AddTag("country", ipInfo.Country_short)
				pt.AddTag("isp", ipInfo.Isp)
			}

			metricsdata = append(metricsdata, []byte(pt.String()))

		} else if IsES(ptname) {

			pipelineInput := map[string]interface{}{}
			pipelineInput["ip"] = sourceIP

			rawFields, _ := pt.Fields()
			for k, v := range rawFields {
				pipelineInput[k] = v
			}

			for _, t := range pt.Tags() {
				pipelineInput[string(t.Key)] = string(t.Value)
			}

			pipelineInputBytes, err := json.Marshal(&pipelineInput)
			if err != nil {
				l.Warnf("%s", err)
				esdata = append(esdata, []byte(pt.String()))
				continue
			}

			l.Debugf("pipeline input: %s", string(pipelineInputBytes))
			pipelineResult, err := pp.Run(string(pipelineInputBytes)).Result()
			if err != nil {
				l.Warnf("%s", err)
				esdata = append(esdata, []byte(pt.String()))
				continue
			} else {
				l.Debugf("pipeline result: %s", pipelineResult)
			}

			tags := influxm.Tags{
				influxm.Tag{
					Key:   []byte("name"),
					Value: []byte(ptname),
				},
			}

			newPt, err := influxm.NewPoint(ptname, tags, pipelineResult, pt.Time())

			if err != nil {
				l.Errorf("%s", err)
				esdata = append(esdata, []byte(pt.String()))
			} else {
				esdata = append(esdata, []byte(newPt.String()))
			}
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
