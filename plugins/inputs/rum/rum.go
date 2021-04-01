package rum

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

const (
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

func (r *Rum) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (r *Rum) RegHttpHandler() {
	l = logger.SLogger(inputName)

	script := r.Pipeline
	if script == "" {
		scriptPath := filepath.Join(datakit.PipelineDir, inputName+".p")
		data, err := ioutil.ReadFile(scriptPath)
		if err == nil {
			script = string(data)
		}
	}

	if script != "" {
		r.pipelinePool = &sync.Pool{
			New: func() interface{} {
				p, err := pipeline.NewPipeline(script)
				if err != nil {
					l.Errorf("%s", err)
				}
				return p
			},
		}
	}

	ipheaderName = r.IPHeader
	httpd.RegGinHandler("POST", io.Rum, r.Handle)
}

func (r *Rum) getPipeline() *pipeline.Pipeline {
	if r.pipelinePool == nil {
		return nil
	}

	pl := r.pipelinePool.Get()
	if pl == nil {
		return nil
	}

	return pl.(*pipeline.Pipeline)
}

func pipelineRUMPoint(pt influxm.Point, pl *pipeline.Pipeline) (influxm.Point, error) {

	if pl == nil {
		return pt, nil
	}

	pl = pl.RunPoint(pt)
	if err := pl.LastError(); err != nil {
		return nil, err
	}

	res, err := pl.Result()
	if err != nil {
		return nil, err
	}

	// XXX: use origin tag
	newpt, err := influxm.NewPoint(string(pt.Name()), pt.Tags(), res, pt.Time())

	if err != nil {
		return nil, err
	}

	return newpt, nil
}

func geoTags(srcip string) (tags map[string]string) {
	tags = map[string]string{}

	ipInfo, err := geo.Geo(srcip)
	if err != nil {
		l.Errorf("geo failed: %s, ignored", err)
		return
	} else {
		// 无脑填充 geo 数据
		tags["city"] = ipInfo.City
		tags["province"] = ipInfo.Region
		tags["country"] = ipInfo.Country_short
		tags["isp"] = ip2isp.SearchIsp(srcip)
	}

	return
}

func (r *Rum) handleBody(body []byte, precision, srcip string) (mpts, rumpts []*influxdb.Point, err error) {
	extraTags := geoTags(srcip)

	mpts, err = lp.ParsePoints(body, &lp.Option{
		ExtraTags: extraTags,
		Strict:    true,
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())
			if isRUMData(name) { // ignore RUM data
				return nil, nil
			}

			if !isMetricData(name) {
				return nil, fmt.Errorf("unknow metric name %s", name)
			}

			return p, nil
		},
	})

	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	pl := r.getPipeline()
	defer func() {
		if pl != nil {
			r.pipelinePool.Put(pl)
		}
	}()
	// add extra ip tag for RUM data
	extraTags["ip"] = srcip

	rumpts, err = lp.ParsePoints(body, &lp.Option{
		ExtraTags: extraTags,
		Strict:    true,
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())
			if isMetricData(name) { // ignore Metric data
				return nil, nil
			}

			if !isRUMData(name) {
				return nil, fmt.Errorf("unknow metric name %s", name)
			}

			newPoint, err := pipelineRUMPoint(p, pl)
			if err != nil { // XXX: ignore pipeline error, use origin point if error
				newPoint = p
			}

			// add extra `message' tag
			newPoint.AddTag("message", newPoint.String())
			return newPoint, nil
		},
	})
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	return mpts, rumpts, nil
}

func (r *Rum) Handle(c *gin.Context) {

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			l.Errorf("panic: %s", err)
			l.Errorf("%s", string(buf[:n]))
		}
	}()

	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error
	srcip := ""

	precision, _ = uhttp.GinGetArg(c, "X-Precision", "precision")

	if ipheaderName != "" {
		srcip = c.Request.Header.Get(ipheaderName)
		if srcip != "" {
			parts := strings.Split(srcip, ",")
			if len(parts) > 0 {
				srcip = parts[0]
			}
		}
	}

	if srcip == "" {
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 0 {
			srcip = parts[0]
		}
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		l.Errorf("%s", err)
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrHttpReadErr, err.Error()))
		return
	}

	metricpts, espts, err := r.handleBody(body, precision, srcip)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
		return
	}

	var x, y []*io.Point
	for _, pt := range metricpts {
		x = append(x, &io.Point{pt})
	}
	for _, pt := range espts {
		y = append(y, &io.Point{pt})
	}

	if len(x) > 0 {
		if err = io.Feed(inputName, io.Metric, x, &io.Option{HighFreq: true}); err != nil {
			uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
			return
		}
	}

	if len(y) > 0 {
		if err = io.Feed(inputName, io.Rum, y, &io.Option{HighFreq: true}); err != nil {
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
