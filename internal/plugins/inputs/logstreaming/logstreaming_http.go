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

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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

func getSourceName(source string) string {
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

const (
	maxLogLen = 32 * 1024 * 1024
)

func (ipt *Input) processLogBody(param *parameters) error {
	var (
		source       = getSourceName(param.queryValues.Get("source"))
		storageIndex = param.queryValues.Get("storage_index")

		// TODO
		// 每一条 request 都要解析和创建一个 tags，影响性能
		// 可以将其缓存起来，以 url 的 md5 值为 key
		extraTags = config.ParseGlobalTags(param.queryValues.Get("tags"))

		urlstr   = param.url.String()
		logPtOpt = point.DefaultLoggingOptions()
		plopt    *lang.LogOption
		pts      []*point.Point
		now      = ntp.Now()
	)

	if !param.ignoreURLTags {
		extraTags["ip_or_hostname"] = param.url.Hostname()
		extraTags["service"] = completeService(source, param.queryValues.Get("service"))
	}

	if scriptName := param.queryValues.Get("pipeline"); scriptName != "" {
		plopt = &lang.LogOption{
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
							kvs = kvs.Add(constants.FieldMessage, v, false, true)
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

					pts = append(pts, point.NewPointV2(source, kvs,
						append(logPtOpt, point.WithExtraTags(extraTags), point.WithTimestamp(ts))...))
				}
				decJSON = true
			}
		}

		if !decJSON {
			var kvs point.KVs
			kvs = kvs.Add(constants.FieldMessage, string(body), false, true)

			pts = append(pts, point.NewPointV2(source, kvs, append(
				logPtOpt,
				point.WithExtraTags(extraTags),
				point.WithTime(now))...),
			)
		}

	default:

		scanBuffer := ipt.getScanbuf()
		defer ipt.putScanbuf(scanBuffer) // nolint:staticcheck

		scanner := bufio.NewScanner(param.body)
		scanner.Buffer(scanBuffer, maxLogLen)

		for scanner.Scan() {
			var kvs point.KVs

			kvs = kvs.AddV2(constants.FieldMessage, scanner.Text(), true).
				AddV2(constants.FieldStatus, pipeline.DefaultStatus, true)
			pts = append(pts, point.NewPointV2(source, kvs,
				append(logPtOpt, point.WithExtraTags(extraTags), point.WithTime(now))...))
		}
	}

	if len(pts) == 0 {
		log.Warnf("len(points) is zero, skip")
		return nil
	}

	feedName := dkio.FeedSource(inputName, source)
	if storageIndex != "" {
		feedName = dkio.FeedSource(feedName, storageIndex)
	}

	return ipt.feeder.Feed(point.Logging, pts,
		dkio.WithSource(feedName),
		dkio.WithStorageIndex(storageIndex),
		dkio.WithPipelineOption(plopt))
}
