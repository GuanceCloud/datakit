// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package httpcli wraps http related functions
package httpcli

import (
	"crypto/tls"
	"net/http/httptrace"
	"runtime"
	"sync"
	"time"
)

var tsPool sync.Pool

func GetTracer(from, remote, url string) *httpClientTraceStat {
	if x := tsPool.Get(); x == nil {
		return newHTTPClientTraceStat(from, remote, url)
	} else {
		s := x.(*httpClientTraceStat)

		s.from = from
		s.remoteAddr = remote
		s.url = url

		s.addTrace()
		return s
	}
}

func putTracer(x *httpClientTraceStat) {
	x.reset()
	tsPool.Put(x)
}

type httpClientTraceStat struct {
	from, remoteAddr, url string

	isReuseConn bool
	isIdle      bool
	idleTime    time.Duration

	dnsStart   int64
	dnsResolve time.Duration

	tlsHSStart int64
	tlsHSDone  time.Duration

	writeBodyStart int64
	writeBodyDone  time.Duration

	connStart int64
	connDone  time.Duration

	wroteRequest int64
	ttfb         time.Duration

	t *httptrace.ClientTrace
}

func (s *httpClientTraceStat) reset() {
	s.from = "-"
	s.remoteAddr = "-"
	s.url = "-"

	s.isReuseConn = false
	s.isIdle = false
	s.idleTime = 0
	s.dnsStart = 0
	s.dnsResolve = 0

	s.tlsHSStart = 0
	s.tlsHSDone = 0

	s.writeBodyStart = 0
	s.writeBodyDone = 0

	s.connStart = 0
	s.connDone = 0
	s.wroteRequest = 0
	s.ttfb = 0
}

func (s *httpClientTraceStat) Trace() *httptrace.ClientTrace {
	return s.t
}

func newHTTPClientTraceStat(from, remote, url string) *httpClientTraceStat {
	s := &httpClientTraceStat{
		from:       from,
		remoteAddr: remote,
		url:        url,
	}
	s.addTrace()
	return s
}

func (s *httpClientTraceStat) Metrics() {
	if s.from == "-" || s.url == "-" {
		_, file, _, ok := runtime.Caller(1)
		if ok {
			s.from = file
		}
	}

	httpClientDNSCost.WithLabelValues(s.from).Observe(s.dnsResolve.Seconds())
	httpClientTLSHandshakeCost.WithLabelValues(s.from).Observe(s.tlsHSDone.Seconds())
	httpClientConnectCost.WithLabelValues(s.from).Observe(s.connDone.Seconds())

	httpClientBodyTransferCost.WithLabelValues(s.from, s.url).Observe(s.writeBodyDone.Seconds())
	httpClientGotFirstResponseByteCost.WithLabelValues(s.from, s.url).Observe(s.ttfb.Seconds())

	httpClientConnIdleTime.WithLabelValues(s.from).Observe(s.idleTime.Seconds())

	if s.isReuseConn {
		httpClientTCPConn.WithLabelValues(s.from, s.remoteAddr, "reused").Add(1)
	} else {
		httpClientTCPConn.WithLabelValues(s.from, s.remoteAddr, "created").Add(1)
	}

	if s.isIdle {
		httpClientConnReusedFromIdle.WithLabelValues(s.from).Add(1)
	}

	putTracer(s)
}

func (s *httpClientTraceStat) addTrace() {
	s.t = &httptrace.ClientTrace{
		GotConn: func(ci httptrace.GotConnInfo) {
			s.isReuseConn = ci.Reused
			s.isIdle = ci.WasIdle
			s.idleTime = ci.IdleTime

			if s.remoteAddr == "" {
				s.remoteAddr = ci.Conn.RemoteAddr().String()
			}
		},

		DNSStart: func(httptrace.DNSStartInfo) { s.dnsStart = time.Now().UnixNano() },
		DNSDone:  func(httptrace.DNSDoneInfo) { s.dnsResolve = time.Since(time.Unix(0, s.dnsStart)) },

		TLSHandshakeStart: func() { s.tlsHSStart = time.Now().UnixNano() },
		TLSHandshakeDone:  func(tls.ConnectionState, error) { s.tlsHSDone = time.Since(time.Unix(0, s.tlsHSStart)) },

		WroteHeaders: func() { s.writeBodyStart = time.Now().UnixNano() },

		WroteRequest: func(_ httptrace.WroteRequestInfo) {
			// NOTE: should we used wrote-request-info here?
			s.wroteRequest = time.Now().UnixNano()
			s.writeBodyDone = time.Since(time.Unix(0, s.writeBodyStart))
		},

		ConnectStart: func(string, string) { s.connStart = time.Now().UnixNano() },
		ConnectDone:  func(string, string, error) { s.connDone = time.Since(time.Unix(0, s.connStart)) },

		GotFirstResponseByte: func() {
			s.ttfb = time.Since(time.Unix(0, s.wroteRequest)) // after wrote request(header + body), then TTFB.
		},
	}
}
