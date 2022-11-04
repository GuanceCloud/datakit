// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package http wraps http related functions
package http

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
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

func GetHeader(req *http.Request, key string) string {
	return strings.ToLower(strings.TrimSpace(req.Header.Get(key)))
}

type NopResponseWriter struct {
	Raw http.ResponseWriter
}

func (nop *NopResponseWriter) Header() http.Header { return make(http.Header) }

func (nop *NopResponseWriter) Write([]byte) (int, error) { return 0, nil }

func (nop *NopResponseWriter) WriteHeader(statusCode int) {}

type HTTPStatusResponse func(resp http.ResponseWriter, req *http.Request, err error)
