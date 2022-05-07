package dataway

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

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

func (dc *endPoint) send(category string, data []byte, gz bool) error {
	var (
		err      error
		isSendOk bool // data sent successfully, http response code is 200
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
			return fmt.Errorf("invalid url %s", category)
		} else {
			log.Debugf("try use URL %+#v", x)
			requrl = category
		}
	}

	var req *http.Request
	if req, err = http.NewRequest("POST", requrl, bytes.NewBuffer(data)); err != nil {
		log.Error(err)

		return err
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

		return nil
	}

	tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header)) //nolint:errcheck,gosec

	var resp *http.Response
	if resp, err = dc.dw.sendReq(req); err != nil {
		dc.fails++
		log.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)

		return err
	}
	span.SetTag("status", resp.Status)

	defer resp.Body.Close() //nolint:errcheck
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error(err)

		return err
	}

	postbeg := time.Now()
	switch resp.StatusCode / 100 {
	case 2:
		isSendOk = true
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

	return err
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
}

func (ts *httpTraceStat) String() string {
	if ts == nil {
		return "-"
	}

	return fmt.Sprintf("dataway httptrace: Conn: [reuse: %v,idle: %v/%s], DNS: %s, TLS: %s, Connect: %s, TTFB: %s",
		ts.reuseConn, ts.idle, ts.idleTime, ts.dnsResolve, ts.tlsHSDone, ts.connDone, ts.ttfbTime)
}

type DatawayErr struct {
	Err   error
	Trace *httpTraceStat
	API   string
}

func (de *DatawayErr) Error() string {
	return fmt.Sprintf("HTTP error: %s, API: %s, httptrace: %s",
		de.Err, de.Trace, de.API)
}

func (dw *DataWayCfg) sendReq(req *http.Request) (*http.Response, error) {
	log.Debugf("send request %s, proxy: %s, dwcli: %p, timeout: %s(%s)",
		req.URL.String(), dw.HTTPProxy, dw.httpCli.Transport,
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
	resp, err := dw.httpCli.Do(req)
	if ts != nil {
		log.Debugf("dataway httptrace: Conn: [reuse: %v,idle: %v/%s], DNS: %s, TLS: %s, Connect: %s, TTFB: %s",
			ts.reuseConn, ts.idle, ts.idleTime, ts.dnsResolve, ts.tlsHSDone, ts.connDone, ts.ttfbTime)
	}

	if err != nil {
		return resp, &DatawayErr{Err: err, Trace: ts, API: req.URL.Path}
	} else {
		return resp, nil
	}

	return resp, err
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {
	for i, ep := range dw.endPoints {
		log.Debugf("send to %dth dataway, fails: %d/%d", i, ep.fails, dw.MaxFails)
		// 判断 fails
		if ep.fails > dw.MaxFails && len(AvailableDataways) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(AvailableDataways)) //nolint:gosec

			var err error
			url := fmt.Sprintf(`%s?%s`, AvailableDataways[index], ep.urlValues.Encode())
			ep, err = dw.initEndpoint(url)
			if err != nil {
				log.Error(err)
				return err
			}

			dw.endPoints[i] = ep
		}

		if err := ep.send(category, data, gz); err != nil {
			return err
		}
	}

	return nil
}
