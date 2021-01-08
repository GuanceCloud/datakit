package http

import (
	"time"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	NAME              = "name"
	PRECISION         = "precision"
	DEFAULT_PRECISION = "ns"
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

	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received metric %d points from %s", len(pts), name)

	// TODO: add global tags

	if err = io.NamedFeed(body, io.Metric, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
	return
}

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

	if err = io.NamedFeed(body, io.Logging, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}

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

	if err = io.NamedFeed(body, io.Tracing, name); err != nil {
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

	if err = io.NamedFeed(body, io.Object, name); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}
