// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

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
	if err = processLogBody(param); err != nil {
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

func processLogBody(param *parameters) error {
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

		var pts []*client.Point
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
		scriptMap := map[string]string{}
		for _, v := range pts1 {
			if v != nil {
				scriptMap[v.Name()] = v.Name() + ".p"
			}
		}
		err = dkio.Feed(inputName, datakit.Logging, pts1, &dkio.Option{PlScript: scriptMap})
	default:
		scanner := bufio.NewScanner(param.body)
		pts := []*point.Point{}
		for scanner.Scan() {
			pt, err := point.NewPoint(source, extraTags,
				map[string]interface{}{
					pipeline.FieldMessage: scanner.Text(),
					pipeline.FieldStatus:  pipeline.DefaultStatus,
				}, point.LOpt())
			if err != nil {
				log.Error(err)
			} else {
				pts = append(pts, pt)
			}
		}
		// pts := plRunCnt(source, pipeLlinePath, pending, extraTags)
		if len(pts) == 0 {
			log.Debugf("len(points) is zero, skip")

			return nil
		}
		err = dkio.Feed(source, datakit.Logging, pts, &dkio.Option{PlScript: map[string]string{source: param.queryValues.Get("pipeline")}})
	}

	return err
}
