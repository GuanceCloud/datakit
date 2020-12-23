package rum

import (
	"bytes"
	"io/ioutil"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"

	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

const (
	PRECISION         = "precision"
	DEFAULT_PRECISION = "ns"
)

var (
	inputName    = `rum`
	moduleLogger *logger.Logger
	ipheaderName = ""
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

func (r *Rum) RegHttpHandler() {
	ipheaderName = r.IPHeader
	moduleLogger = logger.SLogger(inputName)
	httpd.RegGinHandler("POST", io.Rum, Handle)
}

func Handle(c *gin.Context) {

	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error

	contentEncoding := c.Request.Header.Get("Content-Encoding")
	precision = utils.GinGetArg(c, "X-Precision", PRECISION)

	sourceIP := ""

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

	body, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		uhttp.HttpErr(c, err)
		return
	}
	defer c.Request.Body.Close()

	if contentEncoding == "gzip" {
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			uhttp.HttpErr(c, err)
			return
		}
	}

	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
	if err != nil {
		moduleLogger.Errorf("ParsePoints: %s", err.Error())
		e := uhttp.NewErr(err, 400, "")
		e.HttpResp(c)
		return
	}

	moduleLogger.Debugf("received %d points", len(pts))

	metricsdata := [][]byte{}
	esdata := [][]byte{}

	for _, pt := range pts {
		if IsMetric(string(pt.Name())) {
			metricsdata = append(metricsdata, process.NewProcedure(influxdb.NewPointFrom(pt)).Geo(sourceIP).GetByte())
		} else if IsES(string(pt.Name())) {
			esdata = append(esdata, process.NewProcedure(influxdb.NewPointFrom(pt)).Geo(sourceIP).GetByte())
		} else {
			moduleLogger.Warnf("Unsupported rum name: '%s'", string(pt.Name()))
		}
	}

	if len(metricsdata) > 0 {
		body = bytes.Join(metricsdata, []byte("\n"))

		if err = io.NamedFeed(body, io.Metric, inputName); err != nil {
			moduleLogger.Errorf("NamedFeed: %s", err.Error())
			uhttp.HttpErr(c, err)
			return
		}
	}

	if len(esdata) > 0 {
		body = bytes.Join(esdata, []byte("\n"))

		if err = io.NamedFeed(body, io.Rum, inputName); err != nil {
			moduleLogger.Errorf("NamedFeed: %s", err.Error())
			uhttp.HttpErr(c, err)
			return
		}
	}

	utils.ErrOK.HttpResp(c, "ok")
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Rum{}
	})
}
