package http

import (
	"github.com/gin-gonic/gin"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func apiWriteMetric(c *gin.Context) {
	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error

	name := c.Query(NAME)
	precision = c.Query(PRECISION)

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))
		return
	}

	pts, err := lp.ParsePoints(body, &lp.Option{
		ExtraTags: datakit.Cfg.MainCfg.GlobalTags,
		Strict:    true,
		Precision: precision})

	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received metric %d points from %s", len(pts), name)

	var x []*io.Point
	for _, pt := range pts {
		x = append(x, &io.Point{pt})
	}

	if err = io.Feed(name, datakit.Metric, x, nil); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}
