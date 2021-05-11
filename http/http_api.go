package http

import (
	"time"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	NAME              = "name"
	PRECISION         = "precision"
	DEFAULT_PRECISION = "ns"
)

func apiWriteTracing(c *gin.Context) {
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

	l.Debugf("received tracing %d points from %s", len(pts), name)

	// TODO: add global tags

	if err = io.NamedFeed(body, datakit.Tracing, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}

func apiWriteObject(c *gin.Context) {
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

	l.Debugf("received object %d points from %s", len(pts), name)

	// TODO: add global tags

	if err = io.NamedFeed(body, datakit.Object, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}

func apiWriteSecurity(c *gin.Context) {
	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error

	name := c.Query(NAME)
	precision = c.Query(PRECISION)

	if body, err = uhttp.GinRead(c); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))

		return
	}

	var pts []*influxdb.Point
	if pts, err = lp.ParsePoints(body, &lp.Option{
		ExtraTags: datakit.Cfg.MainCfg.GlobalTags,
		Strict:    true,
		Precision: precision}); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))

		return
	}

	l.Debugf("received security %d points from %s", len(pts), name)

	var x []*io.Point
	for _, pt := range pts {
		x = append(x, &io.Point{pt})
	}

	if err = io.Feed(name, datakit.Security, x, nil); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
	} else {
		ErrOK.HttpBody(c, nil)
	}
}

func apiWriteTelegraf(c *gin.Context) {
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("failed to read body, err: %s", err.Error())
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	if len(body) == 0 {
		l.Debug("empty body")
		return
	}

	var pts []*influxdb.Point
	if pts, err = lp.ParsePoints(body, &lp.Option{
		Strict:    true,
		Precision: "n"}); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))

		return
	}

	l.Debugf("received security %d points from %s", len(pts), "telegraf")
	var x []*io.Point
	for _, pt := range pts {
		x = append(x, &io.Point{pt})
	}
	if err = io.Feed("telegraf", datakit.Metric, x, nil); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))

		return
	}
}
