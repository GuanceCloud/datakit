// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/tracer"
)

var startTime = time.Now()

func (dc *endPoint) send(category string, data []byte, gz bool) (int, error) {
	var err error
	var statusCode int
	span, _ := tracer.StartSpanFromContext(context.Background(), "io.dataway.send", tracer.SpanType(ext.SpanTypeHTTP))
	defer func() {
		span.SetTag("fails", dc.fails)
		span.Finish(tracer.WithError(err))
	}()
	span.SetTag("category", category)
	span.SetTag("data_size", len(data))
	span.SetTag("is_gz", gz)

	requrl, ok := dc.categoryURL[category]
	if !ok {
		// for dialtesting, there are user-defined url to post
		if x, err := url.ParseRequestURI(category); err != nil {
			return statusCode, fmt.Errorf("invalid url %s", category)
		} else {
			log.Debugf("try use URL %+#v", x)
			requrl = category
		}
	}

	var req *http.Request
	if req, err = http.NewRequest("POST", requrl, bytes.NewBuffer(data)); err != nil {
		log.Error(err)

		return statusCode, err
	}
	span.SetTag("method", req.Method)
	span.SetTag("url", requrl)

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}
	for k, v := range ExtraHeaders {
		req.Header.Set(k, v)
	}

	if dc.ontest {
		log.Debug("Datakit client on test")

		return statusCode, nil
	}

	tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header)) //nolint:errcheck,gosec

	var resp *http.Response
	if resp, err = dc.dw.sendReq(req); err != nil {
		dc.fails++
		log.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)

		var urlError *url.Error

		if errors.As(err, &urlError) && urlError.Timeout() {
			statusCode = -1 // timeout
		}

		return statusCode, err
	}
	span.SetTag("status", resp.Status)

	defer resp.Body.Close() //nolint:errcheck
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error(err)

		return statusCode, err
	}

	statusCode = resp.StatusCode

	postbeg := time.Now()
	switch statusCode / 100 {
	case 2:
		dc.fails = 0
		log.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(data), requrl, gz, time.Since(postbeg), string(body))
	case 4:
		dc.fails = 0
		log.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(body),
			time.Since(postbeg))
	case 5:
		dc.fails++
		log.Errorf("fails count [%d] post %d to %s failed(HTTP: %s): %s, cost %v", dc.fails,
			len(data),
			requrl,
			resp.Status,
			string(body),
			time.Since(postbeg))
		err = fmt.Errorf("dataway internal error")
	}

	return statusCode, err
}

func (dw *DataWayCfg) sendReq(req *http.Request) (*http.Response, error) {
	log.Debugf("send request %s, proxy: %s, dwcli: %p", req.URL.String(), dw.HTTPProxy, dw.httpCli.Transport)

	return dw.httpCli.Do(req)
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) (statusCode int, err error) {
	for i, ep := range dw.endPoints {
		log.Debugf("send to %dth dataway, fails: %d/%d", i, ep.fails, dw.MaxFails)
		// 判断 fails
		if ep.fails > dw.MaxFails && len(AvailableDataways) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(AvailableDataways)) //nolint:gosec

			url := fmt.Sprintf(`%s?%s`, AvailableDataways[index], ep.urlValues.Encode())
			ep, err = dw.initEndpoint(url)
			if err != nil {
				log.Error(err)
				return
			}

			dw.endPoints[i] = ep
		}

		statusCode, err = ep.send(category, data, gz)
		if err != nil {
			return
		}
	}

	return
}

func (dw *DataWayCfg) Write(category string, pts []sinkcommon.ISinkPoint) error {
	if len(pts) == 0 {
		return nil
	}

	var statusCode int

	bodies, err := dw.buildBody(pts, true)
	if err != nil {
		return err
	}

	rawBytes := 0
	gzBytes := 0

	for _, body := range bodies {
		if code, err := dw.Send(category, body.buf, body.gzon); err != nil {
			log.Error(err)
			sender.FeedMetric(&sender.SinkMetric{
				Name:       "dataway",
				StartTime:  startTime,
				IsSuccess:  false,
				StatusCode: code,
			})
			return err
		} else {
			rawBytes += int(body.rawBufBytes)
			gzBytes += len(body.buf)
			statusCode = code
		}
	}

	sender.FeedMetric(&sender.SinkMetric{
		Name:       "dataway",
		IsSuccess:  true,
		StartTime:  startTime,
		Pts:        uint64(len(pts)),
		Bytes:      uint64(gzBytes),
		RawBytes:   uint64(rawBytes),
		StatusCode: statusCode,
	})

	return nil
}

const (
	minGZSize   = 1024
	maxKodoPack = 10 * 1000 * 1000
)

type body struct {
	buf         []byte
	gzon        bool
	rawBufBytes int64
}

func (dw *DataWayCfg) buildBody(pts []sinkcommon.ISinkPoint, isGzip bool) ([]*body, error) {
	lines := bytes.Buffer{}
	var (
		gz = func(lines []byte) (*body, error) {
			var (
				body = &body{buf: lines, rawBufBytes: int64(len(lines))}
				err  error
			)
			log.Debugf("### io body size before GZ: %dM %dK", len(body.buf)/1000/1000, len(body.buf)/1000)
			if len(lines) > minGZSize && isGzip {
				if body.buf, err = datakit.GZip(body.buf); err != nil {
					log.Errorf("gz: %s", err.Error())

					return nil, err
				}
				body.gzon = true
			}

			return body, nil
		}
		// lines  bytes.Buffer
		bodies []*body
	)
	lines.Reset()
	for _, pt := range pts {
		ptstr := pt.String()
		if lines.Len()+len(ptstr)+1 >= maxKodoPack {
			if body, err := gz(lines.Bytes()); err != nil {
				return nil, err
			} else {
				log.Warn(string(body.buf))
				bodies = append(bodies, body)
			}
			lines.Reset()
		}
		lines.WriteString(ptstr)
		lines.WriteString("\n")
	}
	if body, err := gz(lines.Bytes()); err != nil {
		return nil, err
	} else {
		return append(bodies, body), nil
	}
}
