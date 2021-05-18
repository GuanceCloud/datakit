package http

import (
	"time"

	"github.com/gin-gonic/gin"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func apiWrite(c *gin.Context) {
	var body []byte
	var err error

	category := ""

	if x := c.Param(CATEGORY); x != "" {
		switch x {
		case datakit.Metric,
			datakit.Logging,
			datakit.Object,
			datakit.Tracing,
			datakit.Security,
			datakit.KeyEvent,
			datakit.Rum:
			category = x
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

	input := DEFAULT_INPUT
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

	extraTags := datakit.Cfg.GlobalTags
	if x := c.Query(IGNORE_GLOBAL_TAGS); x != "" {
		extraTags = nil
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))
		return
	}

	if category == datakit.Rum { // RUM 数据单独处理
		handleRUMBody(c, precision, input, body)
		return
	}

	pts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		ExtraTags: extraTags,
		Strict:    true,
		Precision: precision})

	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received %d(%s) points from %s", len(pts), category, input)

	if err = io.Feed(input, category, io.WrapPoint(pts), &io.Option{HighFreq: true}); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}
