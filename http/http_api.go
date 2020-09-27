package http

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
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

	contentEncoding := c.Request.Header.Get("Content-Encoding")
	name := c.Query(NAME)
	precision = c.Query(PRECISION)

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
		l.Errorf("ParsePoints: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	l.Debugf("received %d points from %s", len(pts), name)

	// TODO: add global tags

	if err = io.NamedFeed(body, io.Metric, name); err != nil {
		l.Errorf("NamedFeed: %s", err.Error())
		uhttp.HttpErr(c, err)
	}
	return
}

func apiWriteObject(c *gin.Context) {
	var body []byte
	var err error

	contentEncoding := c.Request.Header.Get("Content-Encoding")
	name := c.Query(NAME)

	body, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		uhttp.HttpErr(c, err)
		return
	}
	defer c.Request.Body.Close()

	if contentEncoding == "gzip" {
		if body, err = utils.ReadCompressed(bytes.NewReader(body), true); err != nil {
			uhttp.HttpErr(c, err)
			return
		}
	}

	if err = io.NamedFeed(body, io.Object, name); err != nil {
		l.Errorf("NamedFeed: %s", err.Error())
		uhttp.HttpErr(c, err)
	}
}
