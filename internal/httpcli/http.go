// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package httpcli wraps http related functions
package httpcli

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

//------------------------------------------------------------------------------
// Options ...

const DefaultKeepAlive = 90 * time.Second

type Options struct {
	DialTimeout   time.Duration
	DialKeepAlive time.Duration

	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	InsecureSkipVerify    bool
	ProxyURL              *url.URL
	DialContext           dnet.DialFunc
}

func NewOptions() *Options {
	return &Options{
		DialTimeout:           30 * time.Second,
		DialKeepAlive:         DefaultKeepAlive,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   runtime.NumGoroutine(),
		IdleConnTimeout:       DefaultKeepAlive, // keep the same with keep-aliva
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		DialContext:           nil,
	}
}

//------------------------------------------------------------------------------

func DefTransport() *http.Transport {
	return newCliTransport(NewOptions())
}

//nolint:gomnd
func newCliTransport(opt *Options) *http.Transport {
	var (
		proxy       func(*http.Request) (*url.URL, error)
		dialContext dnet.DialFunc
	)

	if opt.ProxyURL != nil {
		proxy = http.ProxyURL(opt.ProxyURL)
	}

	if opt.DialContext != nil {
		dialContext = opt.DialContext
	} else {
		dialContext = (&net.Dialer{
			Timeout: func() time.Duration {
				if opt.DialTimeout > time.Duration(0) {
					return opt.DialTimeout
				}
				return 30 * time.Second
			}(),
			KeepAlive: func() time.Duration {
				if opt.DialKeepAlive > time.Duration(0) {
					return opt.DialKeepAlive
				}
				return DefaultKeepAlive
			}(),
		}).DialContext
	}

	return &http.Transport{
		Proxy:       proxy,
		DialContext: dialContext,

		MaxIdleConns: func() int {
			if opt.MaxIdleConns == 0 {
				return 100
			}
			return opt.MaxIdleConns
		}(),

		TLSClientConfig: func() *tls.Config {
			if opt.InsecureSkipVerify {
				return &tls.Config{InsecureSkipVerify: true} //nolint:gosec
			}
			return &tls.Config{InsecureSkipVerify: false} //nolint:gosec
		}(),

		MaxIdleConnsPerHost: func() int {
			if opt.MaxIdleConnsPerHost == 0 {
				return runtime.NumGoroutine()
			}
			return opt.MaxIdleConnsPerHost
		}(),

		IdleConnTimeout: func() time.Duration {
			if opt.IdleConnTimeout > time.Duration(0) {
				return opt.IdleConnTimeout
			}
			return DefaultKeepAlive
		}(),

		TLSHandshakeTimeout: func() time.Duration {
			if opt.TLSHandshakeTimeout > time.Duration(0) {
				return opt.TLSHandshakeTimeout
			}
			return 10 * time.Second
		}(),

		ExpectContinueTimeout: func() time.Duration {
			if opt.ExpectContinueTimeout > time.Duration(0) {
				return opt.ExpectContinueTimeout
			}
			return time.Second
		}(),
	}
}

func Cli(opt *Options) *http.Client {
	if opt == nil {
		return &http.Client{
			Transport: DefTransport(),
		}
	}

	return &http.Client{
		Transport: newCliTransport(opt),
	}
}

func RemoteAddr(req *http.Request) (ip, port string) {
BREAKPOINT:
	for _, h := range []string{
		"x-forwarded-for",
		"X-FORWARDED-FOR",
		"X-Forwarded-For",
		"x-real-ip",
		"X-REAL-IP",
		"X-Real-Ip",
		"proxy-client-ip",
		"PROXY-CLIENT-IP",
		"Proxy-Client-Ip",
	} {
		addrs := strings.Split(req.Header.Get(h), ",")
		for _, addr := range addrs {
			if ip, port, _ = net.SplitHostPort(addr); ip == "" {
				continue
			}
			break BREAKPOINT
		}
	}
	if ip == "" {
		ip, port, _ = net.SplitHostPort(req.RemoteAddr)
	}

	return
}

type ConnWatcher struct {
	New      int64
	Close    int64
	Max      int64
	Idle     int64
	Active   int64
	Hijacked int64
}

func (cw *ConnWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt64(&cw.New, 1)
		atomic.AddInt64(&cw.Max, 1)

	case http.StateHijacked:
		atomic.AddInt64(&cw.Hijacked, 1)
		atomic.AddInt64(&cw.New, -1)

	case http.StateClosed:
		atomic.AddInt64(&cw.New, -1)
		atomic.AddInt64(&cw.Close, 1)

	case http.StateIdle:
		atomic.AddInt64(&cw.Idle, 1)

	case http.StateActive:
		atomic.AddInt64(&cw.Active, 1)
	}
}

func (cw *ConnWatcher) Count() int {
	return int(atomic.LoadInt64(&cw.New))
}

func (cw *ConnWatcher) String() string {
	return fmt.Sprintf("connections: new: %d, closed: %d, Max: %d, hijacked: %d, idle: %d, active: %d",
		cw.New, cw.Close, cw.Max, cw.Hijacked, cw.Idle, cw.Active)
}

type httpClientTraceStat struct {
	from string

	reuseConn bool
	idle      bool
	idleTime  time.Duration

	remoteAddr string

	dnsStart   time.Time
	dnsResolve time.Duration
	tlsHSStart time.Time
	tlsHSDone  time.Duration
	connStart  time.Time
	connDone   time.Duration

	wroteRequest time.Time
	ttfb         time.Duration

	t *httptrace.ClientTrace
}

func (s *httpClientTraceStat) Trace() *httptrace.ClientTrace {
	return s.t
}

// NewHTTPClientTraceStat create a hook for HTTP client running metrics.
func NewHTTPClientTraceStat(from string) *httpClientTraceStat {
	s := &httpClientTraceStat{
		from: from,
	}

	s.addTrace()

	return s
}

func (s *httpClientTraceStat) Metrics() {
	httpClientDNSCost.WithLabelValues(s.from).Observe(float64(s.dnsResolve) / float64(time.Second))
	httpClientTLSHandshakeCost.WithLabelValues(s.from).Observe(float64(s.tlsHSDone) / float64(time.Second))
	httpClientConnectCost.WithLabelValues(s.from).Observe(float64(s.connDone) / float64(time.Second))
	httpClientGotFirstResponseByteCost.WithLabelValues(s.from).Observe(float64(s.ttfb) / float64(time.Second))

	httpClientConnIdleTime.WithLabelValues(s.from).Observe(float64(s.idleTime) / float64(time.Second))
	if s.reuseConn {
		httpClientTCPConn.WithLabelValues(s.from, s.remoteAddr, "reused").Add(1)
	} else {
		httpClientTCPConn.WithLabelValues(s.from, s.remoteAddr, "created").Add(1)
	}

	if s.idle {
		httpClientConnReusedFromIdle.WithLabelValues(s.from).Add(1)
	}
}

func (s *httpClientTraceStat) addTrace() {
	s.t = &httptrace.ClientTrace{
		GotConn: func(ci httptrace.GotConnInfo) {
			s.reuseConn = ci.Reused
			s.idle = ci.WasIdle
			s.idleTime = ci.IdleTime
			s.remoteAddr = ci.Conn.RemoteAddr().String()
		},

		DNSStart: func(httptrace.DNSStartInfo) { s.dnsStart = time.Now() },
		DNSDone:  func(httptrace.DNSDoneInfo) { s.dnsResolve = time.Since(s.dnsStart) },

		TLSHandshakeStart: func() { s.tlsHSStart = time.Now() },
		TLSHandshakeDone:  func(tls.ConnectionState, error) { s.tlsHSDone = time.Since(s.tlsHSStart) },

		WroteRequest: func(_ httptrace.WroteRequestInfo) {
			// NOTE: should we used wrote-request-info here?
			s.wroteRequest = time.Now()
		},

		ConnectStart: func(string, string) { s.connStart = time.Now() },
		ConnectDone:  func(string, string, error) { s.connDone = time.Since(s.connStart) },

		GotFirstResponseByte: func() {
			s.ttfb = time.Since(s.wroteRequest) // after wrote request(header + body), then TTFB.
		},
	}
}
