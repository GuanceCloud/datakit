package io

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var (
	errEmptyBody = uhttp.NewErr(errors.New("empty body"), http.StatusBadRequest, "datakit")
	httpOK       = uhttp.NewErr(nil, http.StatusOK, "datakit")
)

func HandleTelegrafOutput(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		l.Errorf("read http body failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	defer c.Request.Body.Close()

	if len(body) == 0 {
		l.Errorf("read http body failed: %s", err.Error())
		uhttp.HttpErr(c, errEmptyBody)
		return
	}

	// NOTE:
	// - we only accept nano-second precison here
	// - no gzip content-encoding support
	// - only accept influx line protocol
	// so be careful to apply telegraf http output

	points, err := models.ParsePointsWithPrecision(body, time.Now().UTC(), "n")
	feeds := map[string][]string{}

	for _, p := range points {
		meas := string(p.Name())
		if _, ok := feeds[meas]; !ok {
			feeds[meas] = []string{}
		}

		feeds[meas] = append(feeds[meas], p.String())
	}

	for k, lines := range feeds {
		if err := NamedFeed([]byte(strings.Join(lines, "\n")), Metric, k); err != nil {
			uhttp.HttpErr(c, err)
			return
		}
	}

	httpOK.HttpResp(c)
}
