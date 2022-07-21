// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/tracer"
)

var (
	sendFailStats = map[string]int32{}
	lock          sync.RWMutex
)

// GetSendStat return the sent fail count of the specified category.
func GetSendStat(category string) int32 {
	lock.RLock()
	defer lock.RUnlock()
	if failCount, ok := sendFailStats[category]; ok {
		return failCount
	} else {
		return 0
	}
}

func updateSendFailStats(category string, isOk bool) {
	lock.Lock()
	defer lock.Unlock()

	if failCount, ok := sendFailStats[category]; ok {
		if isOk {
			sendFailStats[category] = 0
		} else {
			sendFailStats[category] = failCount + 1
		}
	} else {
		if !isOk {
			sendFailStats[category] = 1
		}
	}

	log.Debugf("update send fail stats: %+#v", sendFailStats)
}

func (dc *endPoint) send(category string, data []byte, gz bool) (int, error) {
	var (
		err        error
		isSendOk   bool // data sent successfully, http response code is 200
		statusCode int
	)

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
		// update send stats
		defer func() {
			updateSendFailStats(category, isSendOk)
		}()

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

	var (
		resp    *http.Response
		postbeg = time.Now()
	)
	if resp, err = dc.dw.sendReq(req); err != nil {
		dc.fails++
		log.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)

		if dwError, ok := err.(*DatawayError); ok { //nolint:errorlint
			err := dwError.Err
			var urlError *url.Error
			if errors.As(err, &urlError) && urlError.Timeout() {
				statusCode = -1 // timeout
			}
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

	switch resp.StatusCode / 100 {
	case 2:
		isSendOk = true
		dc.fails = 0
		log.Debugf("post %d to %s ok(gz: %v), cost %v", len(data), requrl, gz, time.Since(postbeg))
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

type httpTraceStat struct {
	reuseConn bool
	idle      bool
	idleTime  time.Duration

	dnsStart   time.Time
	dnsResolve time.Duration
	tlsHSStart time.Time
	tlsHSDone  time.Duration
	connStart  time.Time
	connDone   time.Duration
	ttfbTime   time.Duration

	cost time.Duration
}

func (ts *httpTraceStat) String() string {
	if ts == nil {
		return "-"
	}

	return fmt.Sprintf("dataway httptrace: Conn: [reuse: %v,idle: %v/%s], DNS: %s, TLS: %s, Connect: %s, TTFB: %s, cost: %s",
		ts.reuseConn, ts.idle, ts.idleTime, ts.dnsResolve, ts.tlsHSDone, ts.connDone, ts.ttfbTime, ts.cost)
}

type DatawayError struct {
	Err   error
	Trace *httpTraceStat
	API   string
}

func (de *DatawayError) Error() string {
	return fmt.Sprintf("HTTP error: %s, API: %s, httptrace: %s",
		de.Err, de.API, de.Trace)
}

func (dw *DataWayDefault) sendReq(req *http.Request) (*http.Response, error) {
	log.Debugf("send request %s, proxy: %s, dwcli: %p, timeout: %s(%s)",
		req.URL.String(), dw.HTTPProxy, dw.httpCli.HTTPClient.Transport,
		dw.HTTPTimeout, dw.TimeoutDuration.String())

	var reqStart time.Time
	var ts *httpTraceStat
	if dw.EnableHTTPTrace {
		ts = &httpTraceStat{}
		t := &httptrace.ClientTrace{
			GotConn: func(ci httptrace.GotConnInfo) {
				ts.reuseConn = ci.Reused
				ts.idle = ci.WasIdle
				ts.idleTime = ci.IdleTime
			},
			DNSStart:             func(httptrace.DNSStartInfo) { ts.dnsStart = time.Now() },
			DNSDone:              func(httptrace.DNSDoneInfo) { ts.dnsResolve = time.Since(ts.dnsStart) },
			TLSHandshakeStart:    func() { ts.tlsHSStart = time.Now() },
			TLSHandshakeDone:     func(tls.ConnectionState, error) { ts.tlsHSDone = time.Since(ts.tlsHSStart) },
			ConnectStart:         func(string, string) { ts.connStart = time.Now() },
			ConnectDone:          func(string, string, error) { ts.connDone = time.Since(ts.connStart) },
			GotFirstResponseByte: func() { ts.ttfbTime = time.Since(reqStart) },
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), t))
	}

	reqStart = time.Now()
	x, err := retryablehttp.FromRequest(req)
	if err != nil {
		log.Errorf("retryablehttp.FromRequest: %s", err)
		return nil, err
	}

	dw.locker.Lock()
	resp, err := dw.httpCli.Do(x)
	dw.locker.Unlock()
	if ts != nil {
		ts.cost = time.Since(reqStart)
		log.Debugf("%s: %s", req.URL.Path, ts.String())
	}

	if err != nil {
		return nil, &DatawayError{Err: err, Trace: ts, API: req.URL.Path}
	}

	return resp, nil
}

func (dw *DataWayDefault) Send(category string, data []byte, gz bool) (statusCode int, err error) {
	for i, ep := range dw.endPoints {
		log.Debugf("send to %dth dataway, fails: %d/%d", i, ep.fails, dw.MaxFails)
		// 判断 fails
		if ep.fails > dw.MaxFails && len(AvailableDataways) > 0 {
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

func (dw *DataWayDefault) Write(category string, pts []*point.Point) (*sinkcommon.Failed, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	bodies, err := buildBody(pts, true)
	if err != nil {
		return nil, err
	}

	var failed *sinkcommon.Failed

	log.Debugf("write %d pts (%d parts) to %s", len(pts), len(bodies), category)

	for idx, body := range bodies {
		log.Debugf("write %dth part(%d bytes, gz: %v, raw: %d) to %s",
			idx, len(body.buf), body.gzon, body.rawBufBytes, category)
		if _, err := dw.Send(category, body.buf, body.gzon); err != nil {
			if failed == nil {
				failed = &sinkcommon.Failed{}
			}
			failed.Ranges = append(failed.Ranges, body.idxRange)

			continue
		}
	}

	return failed, nil
}

const (
	minGZSize   = 1024
	maxKodoPack = 10 * 1000 * 1000
)

type body struct {
	buf         []byte
	gzon        bool
	rawBufBytes int64
	idxRange    [2]int
}

func buildBody(pts []*point.Point, isGzip bool) ([]*body, error) {
	lines := bytes.Buffer{}

	var (
		getBody = func(lines []byte, idxBegin, idxEnd int) (*body, error) {
			var (
				body = &body{
					buf:         lines,
					rawBufBytes: int64(len(lines)),
					idxRange:    [2]int{idxBegin, idxEnd},
				}
				err error
			)

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

	bodyIdx := 0
	for idx, pt := range pts {
		ptstr := pt.String()
		if lines.Len()+len(ptstr)+1 >= maxKodoPack {
			if body, err := getBody(lines.Bytes(), bodyIdx, idx); err != nil {
				return nil, err
			} else {
				bodyIdx = idx
				bodies = append(bodies, body)
			}
			lines.Reset()
		}
		lines.WriteString(ptstr)
		lines.WriteString("\n")
	}

	if body, err := getBody(lines.Bytes(), bodyIdx, len(pts)); err != nil {
		return nil, err
	} else {
		return append(bodies, body), nil
	}
}
