// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming http handler.
package logstreaming

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
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

	return defaultMeasurementName
}

func completeService(defaultService, service string) string {
	if service != "" {
		return service
	}

	return defaultService
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
		urlstr   = param.url.String()
		logPtOpt = point.DefaultLoggingOptions()
		plopt    *plmanager.Option
		pts      []*point.Point
		now      = time.Now()
	)

	if scriptName := param.queryValues.Get("pipeline"); scriptName != "" {
		plopt = &plmanager.Option{
			ScriptMap: map[string]string{
				source: scriptName,
			},
		}
	}

	switch param.queryValues.Get("type") {
	case "influxdb":
		body, err := ioutil.ReadAll(param.body)
		if err != nil {
			log.Errorf("url %s failed to read body: %s", urlstr, err)
			return err
		}

		dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)

		pts, err = dec.Decode(body,
			append(logPtOpt,
				point.WithExtraTags(extraTags),
				point.WithPrecision(point.PrecStr(param.queryValues.Get("precision"))),
			)...)
		if err != nil {
			log.Errorf("url %s handler err: %s", urlstr, err)
			return err
		}

	case "firelens":

		body, err := ioutil.ReadAll(param.body)
		if err != nil {
			log.Errorf("url %s failed to read body: %s", urlstr, err)
			return err
		}

		var decJSON bool
		if json.Valid(body) {
			var arr []map[string]any
			if err := json.Unmarshal(body, &arr); err != nil {
				log.Debug("url %s, json unmarshal err: %s", urlstr, err)
			} else {
				for _, v := range arr {
					kvs := make(point.KVs, 0, len(v))
					var ts int64
					for k, v := range v {
						switch k {
						case "log":
							kvs = kvs.Add(pipeline.FieldMessage, v, false, true)
						case "date":
							switch v := v.(type) {
							case float64:
								ts = int64(v * 1e9)
							case float32:
								ts = int64(v * 1e9)
							case int64:
								ts = v * 1e9
							case int32:
								ts = int64(v) * 1e9
							default:
								kvs = kvs.Add(k, v, false, true)
							}
						case "source":
							kvs = kvs.Add("firelens_source", v, false, true)
						default:
							kvs = kvs.Add(k, v, false, true)
						}
					}

					if ts == 0 {
						ts = now.UnixNano()
					}

					pt := point.NewPointV2(source, kvs, append(
						logPtOpt,
						point.WithExtraTags(extraTags),
						point.WithTime(time.Unix(0, ts)))...)
					pts = append(pts, pt)
				}
				decJSON = true
			}
		}

		if !decJSON {
			var kvs point.KVs
			kvs = kvs.Add(pipeline.FieldMessage, string(body), false, true)

			pts = append(pts, point.NewPointV2(source, kvs, append(
				logPtOpt,
				point.WithExtraTags(extraTags),
				point.WithTime(now))...),
			)
		}

	default:
		scanner := bufio.NewScanner(param.body)
		for scanner.Scan() {
			var kvs point.KVs

			kvs = kvs.Add(pipeline.FieldMessage, scanner.Text(), false, true)
			kvs = kvs.Add(pipeline.FieldStatus, pipeline.DefaultStatus, false, true)
			pt := point.NewPointV2(source, kvs,
				append(logPtOpt, point.WithExtraTags(extraTags), point.WithTime(now))...)
			pts = append(pts, pt)
		}
	}

	if len(pts) == 0 {
		log.Debugf("len(points) is zero, skip")
		return nil
	}

	return ipt.feeder.FeedV2(point.Logging, pts,
		dkio.WithInputName(inputName),
		dkio.WithPipelineOption(plopt))
}
