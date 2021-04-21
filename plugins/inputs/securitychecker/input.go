package securitychecker

import (
	"github.com/gin-gonic/gin"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "securitychecker"
	sampleCfg = ``
	l         = logger.SLogger(inputName)
)

type Input struct {
}

func (i *Input) Catalog() string { return inputName }

func (i *Input) SampleConfig() string { return sampleCfg }

func (i *Input) Run() { l.Infof("%s input started...", inputName) }

func (i *Input) RegHttpHandler() {
	httpd.RegGinHandler("POST", io.Security, i.handler)
}

func (i *Input) handler(ctx *gin.Context) {
	l.Infof("get input from %d", inputName)

	var (
		precision = ctx.Query(httpd.PRECISION)
		body      []byte
		err       error
	)
	if precision == "" {
		precision = httpd.DEFAULT_PRECISION
	}

	if body, err = uhttp.GinRead(ctx); err != nil {
		l.Error(err.Error())
		uhttp.HttpErr(ctx, uhttp.Error(httpd.ErrHttpReadErr, err.Error()))

		return
	}

	var points []*influxdb.Point
	points, err = lp.ParsePoints(body, &lp.Option{
		Precision: precision,
		ExtraTags: map[string]string{
			"category": "",
			"host":     "",
			"level":    "",
			"title":    "",
		},
		Strict: true,
	})
	if err != nil {
		l.Error(err.Error())
		uhttp.HttpErr(ctx, uhttp.Error(httpd.ErrBadReq, err.Error()))

		return
	}

	l.Info(points)

	// pts := make([]*io.Point, len(points))
}

func init() {
	inputs.Add(inputName, func() inputs.Input { return &Input{} })
}
