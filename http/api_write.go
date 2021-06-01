package http

import (
	"time"

	"github.com/gin-gonic/gin"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func apiWrite(c *gin.Context) {
	var body []byte
	var err error

	input := DEFAULT_INPUT

	category := ""

	if x := c.Param(CATEGORY); x != "" {
		switch x {
		case datakit.Metric,
			datakit.Logging,
			datakit.Object,
			datakit.Tracing,
			datakit.KeyEvent:

			category = x
		case datakit.Rum:
			category = x
			input = "rum"
		case datakit.Security:
			category = x
			input = "sechecker"
		default:
			l.Debugf("invalid category: %s", x)
			uhttp.HttpErr(c, ErrInvalidCategory)
			return
		}
	} else {
		l.Debug("empty category")
		uhttp.HttpErr(c, ErrInvalidCategory)
		return
	}

	if x := c.Query(INPUT); x != "" {
		input = x
	}

	precision := DEFAULT_PRECISION
	if x := c.Query(PRECISION); x != "" {
		precision = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", precision)
		uhttp.HttpErr(c, ErrInvalidPrecision)
		return
	}

	extraTags := config.Cfg.GlobalTags
	if x := c.Query(IGNORE_GLOBAL_TAGS); x != "" {
		extraTags = nil
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))
		return
	}

	l.Debugf("body: %s", string(body))

	if category == datakit.Rum { // RUM 数据单独处理
		handleRUM(c, precision, input, body)
		return
	}

	pts, err := handleWriteBody(body, extraTags, precision)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received %d(%s) points from %s", len(pts), category, input)

	err = io.Feed(input, category, io.WrapPoint(pts), &io.Option{HighFreq: true})

	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
	} else {
		ErrOK.HttpBody(c, nil)
	}
}

func handleWriteBody(body []byte, tags map[string]string, precision string) (pts []*influxdb.Point, err error) {

	pts, err = lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		ExtraTags: tags,
		Strict:    true,
		Precision: precision})

	if err != nil {
		return nil, err
	}

	return
}
