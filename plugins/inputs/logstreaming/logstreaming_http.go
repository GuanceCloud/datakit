// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming http handler.
package logstreaming

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	"github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	resp.Write([]byte(`{"status":"success"}`)) // nolint: errcheck,gosec
}

type parameters struct {
	ignoreURLTags bool
	url           *url.URL
	queryValues   url.Values
	body          *bytes.Buffer
}

func (ipt *Input) handleLogstreaming(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### Log request from %s", req.URL.String())

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &parameters{
		ignoreURLTags: ipt.IgnoreURLTags,
		url:           req.URL,
		queryValues:   req.URL.Query(),
		body:          pbuf,
	}
	if err = ipt.processLogBody(param); err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	resp.Write([]byte(`{"status":"success"}`)) //nolint:errcheck,gosec
}

func completeSource(source string) string {
	if source != "" {
		return source
	}

	return "default"
}

func completeService(defaultService, service string) string {
	if service != "" {
		return service
	}

	return defaultService
}

func completePrecision(precision string) string {
	if precision == "" {
		return defaultPercision
	}

	return precision
}

func (ipt *Input) processLogBody(param *parameters) error {
	var (
		source = completeSource(param.queryValues.Get("source"))
		// TODO
		// 每一条 request 都要解析和创建一个 tags，影响性能
		// 可以将其缓存起来，以 url 的 md5 值为 key
		extraTags = config.ParseGlobalTags(param.queryValues.Get("tags"))
	)
	if !param.ignoreURLTags {
		extraTags["ip_or_hostname"] = param.url.Hostname()
		extraTags["service"] = completeService(source, param.queryValues.Get("service"))
	}

	var (
		urlstr = param.url.String()
		err    error
	)
	switch param.queryValues.Get("type") {
	case "influxdb":
		body, _err := ioutil.ReadAll(param.body)
		if _err != nil {
			log.Errorf("url %s failed to read body: %s", urlstr, _err)
			err = _err
			break
		}

		var pts []*influxdb.Point
		if pts, err = lp.ParsePoints(body, &lp.Option{
			Time:      time.Now(),
			ExtraTags: extraTags,
			Strict:    true,
			Precision: completePrecision(param.queryValues.Get("precision")),
		}); err != nil {
			log.Errorf("url %s handler err: %s", urlstr, err)

			return err
		}
		if len(pts) == 0 {
			log.Debugf("len(points) is zero, skip")

			return nil
		}

		pts1 := point.WrapPoint(pts)
		err = ipt.feeder.Feed(inputName, point.Logging, pts1, nil)

	default:
		scanner := bufio.NewScanner(param.body)
		pts := make([]*point.Point, 0)
		opts := point.DefaultLoggingOptions()
		for scanner.Scan() {
			fields := map[string]interface{}{
				pipeline.FieldMessage: scanner.Text(),
				pipeline.FieldStatus:  pipeline.DefaultStatus,
			}
			pts = append(pts, point.NewPointV2([]byte(source), append(point.NewTags(extraTags), point.NewKVs(fields)...), opts...))
		}

		if len(pts) == 0 {
			log.Debugf("len(points) is zero, skip")

			return nil
		}

		var scriptMap map[string]string

		if scriptName := param.queryValues.Get("pipeline"); scriptName != "" {
			scriptMap = map[string]string{
				source: scriptName,
			}
		}

		err = ipt.feeder.Feed(source, point.Logging, pts, &dkio.Option{PlScript: scriptMap})
	}

	return err
}
