// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var (
	sendFailStats = map[string]int32{}
	lock          sync.RWMutex
	seprator      = []byte("\n")

	maxKodoPack = uint64(10 * 1000 * 1000)
	minGZSize   = 1024
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
		log.Debugf("post %d to %s ok(gz: %v), cost %v", len(data), requrl, gz, time.Since(postbeg))
	case 4:
		log.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(body),
			time.Since(postbeg))
	case 5:
		log.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
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
	for _, ep := range dw.endPoints {
		statusCode, err = ep.send(category, data, gz)
		if err != nil {
			return
		}
	}

	return
}

func (dw *DataWayDefault) Write(category string, pts []*point.Point) (*point.Failed, error) {
	start := time.Now()

	if len(pts) == 0 {
		return nil, nil
	}

	bodies, err := buildBody(pts, true)
	if err != nil {
		return nil, err
	}

	var failed *point.Failed
	log.Debugf("building %d pts to %s, cost: %s", len(pts), category, time.Since(start))

	log.Debugf("write %d pts (%d parts) to %s", len(pts), len(bodies), category)

	start = time.Now()
	for idx, body := range bodies {
		log.Debugf("write %dth part(%d bytes, gz: %v, raw: %d) to %s",
			idx, len(body.buf), body.gzon, body.rawBufBytes, category)
		if _, err := dw.Send(category, body.buf, body.gzon); err != nil {
			if failed == nil {
				failed = &point.Failed{}
			}
			failed.Ranges = append(failed.Ranges, body.idxRange)

			continue
		}
	}

	log.Debugf("sending %d pts to %s, cost: %s", len(pts), category, time.Since(start))

	return failed, nil
}

type body struct {
	buf         []byte
	gzon        bool
	rawBufBytes int64
	idxRange    [2]int
}

func (b *body) String() string {
	return fmt.Sprintf("gzon: %v, range: %v, raw bytes: %d, buf bytes: %d",
		b.gzon, b.idxRange, b.rawBufBytes, len(b.buf))
}

func buildBody(pts []*point.Point, isGzip bool) ([]*body, error) {
	lines := [][]byte{}
	curPartSize := 0

	getBody := func(lines [][]byte, idxBegin, idxEnd int) (*body, error) {
		var (
			body = &body{
				buf:      bytes.Join(lines, seprator),
				idxRange: [2]int{idxBegin, idxEnd},
			}
			err error
		)

		body.rawBufBytes = int64(len(body.buf))

		if curPartSize >= minGZSize && isGzip {
			if body.buf, err = datakit.GZip(body.buf); err != nil {
				log.Errorf("gz: %s", err.Error())

				return nil, err
			}
			body.gzon = true
		}

		return body, nil
	}

	var (
		bodies []*body
	)

	idxBegin := 0
	for idx, pt := range pts {
		ptbytes := []byte(pt.String())

		// 此处必须提前预判包是否会大于上限值，当新进来的 ptbytes 可能
		// 会超过上限时，就应该及时将已有数据（肯定没超限）打包一下。
		if uint64(curPartSize+len(lines)+len(ptbytes)) >= maxKodoPack {
			log.Debugf("merge %d points as body", len(lines))

			if body, err := getBody(lines, idxBegin, idx); err != nil {
				return nil, err
			} else {
				idxBegin = idx
				bodies = append(bodies, body)
				lines = lines[:0]
				curPartSize = 0
			}
		}

		// 如果上面有打包，这里将是一个新的包，否则 ptbytes 还是追加到
		// 已有数据上。
		lines = append(lines, ptbytes)
		curPartSize += len(ptbytes)
	}

	// TODO: Huge, only for tester, if go to production, comment this.
	// log.Debugf("lines to send: %s", lines.String())

	if len(lines) > 0 { // 尾部 lines 单独打包一下
		if body, err := getBody(lines, idxBegin, len(pts)); err != nil {
			return nil, err
		} else {
			return append(bodies, body), nil
		}
	} else {
		return bodies, nil
	}
}
