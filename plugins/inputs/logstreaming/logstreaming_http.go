// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
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
	log.Debugf("### received log request from %s", req.URL.String())

	var (
		readbodycost = time.Now()
		enterWPoolOK = false
	)
	pbuf := bufpool.GetBuffer()
	defer func() {
		if !enterWPoolOK {
			bufpool.PutBuffer(pbuf)
		}
	}()

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

	log.Debugf("### path: %s, body-data-type: %s, body-size: %d, read-body-cost: %dms",
		param.url.Path, param.queryValues.Get("type"), pbuf.Len(), time.Since(readbodycost)/time.Millisecond)

	if wpool == nil {
		if err = processLogBody(param); err != nil {
			log.Error(err.Error())
			resp.Write([]byte(fmt.Sprintf(`{"status":"fail","error_msg":%q}`, err.Error()))) //nolint:errcheck,gosec

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(param),
			workerpool.WithProcess(processLogBodyAdapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration) {
				if param, ok := input.(*parameters); ok {
					bufpool.PutBuffer(param.body)
				}
				log.Debugf("### job status: input: %v, output: %v, cost: %dms", input, output, cost/time.Millisecond)
			}),
		)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		if err = wpool.MoreJob(job); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusTooManyRequests)

			return
		}
		enterWPoolOK = true
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

func processLogBodyAdapter(input interface{}) (output interface{}) {
	param, ok := input.(*parameters)
	if !ok {
		return errors.New("type assertion failed")
	}

	return processLogBody(param)
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

		var pts []*lp.Point

		precisionV2, _err := lp.ConvertPrecisionToV2(completePrecision(param.queryValues.Get("precision")))
		if _err != nil {
			return fmt.Errorf("not support precision: %w", err)
		}

		if pts, err = lp.ParseWithOptionSetter(body,
			lp.WithTime(time.Now()),
			lp.WithExtraTags(extraTags),
			lp.WithStrict(true),
			lp.WithPrecisionV2(precisionV2),
		); err != nil {
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
				scriptMap[v.Name] = v.Name + ".p"
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
