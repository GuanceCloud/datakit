package http

import (
	"github.com/gin-gonic/gin"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func apiWriteLogging(c *gin.Context) {
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

	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received logging %d points from %s", len(pts), name)

	// TODO: add global tags

	if err = io.NamedFeed(body, datakit.Logging, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}
