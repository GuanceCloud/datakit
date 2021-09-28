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
)

func DefTransport() *http.Transport {
	return newCliTransport(&Options{
		DialTimeout:           30 * time.Second,
		DialKeepAlive:         30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   runtime.NumGoroutine(),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	})
}

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
}

//nolint:gomnd
func newCliTransport(opt *Options) *http.Transport {
	var proxy func(*http.Request) (*url.URL, error)

	if opt.ProxyURL != nil {
		proxy = http.ProxyURL(opt.ProxyURL)
	}

	return &http.Transport{
		DialContext: (&net.Dialer{
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
				return 30 * time.Second
			}(),
		}).DialContext,

		Proxy: proxy,

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
			return 90 * time.Second
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
