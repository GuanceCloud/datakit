package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

var (
	defTransport = cliTransport(&Options{
		DialTimeout:           30 * time.Second,
		DialKeepAlive:         30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   runtime.NumGoroutine(),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	})
)

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
func cliTransport(opt *Options) *http.Transport {
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
			Transport: defTransport,
		}
	}

	return &http.Client{
		Transport: cliTransport(opt),
	}
}

func SendRequest(req *http.Request) (*http.Response, error) {
	return (&http.Client{Transport: defTransport}).Do(req)
}

func SendRequestWithTimeout(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{
		Transport: defTransport,
		Timeout:   timeout,
	}).Do(req)
}

func RemoteAddr(req *http.Request) (ip, port string) {
BREAKPOINT:
	for _, h := range []string{"x-forwarded-for",
		"X-FORWARDED-FOR",
		"X-Forwarded-For",
		"x-real-ip",
		"X-REAL-IP",
		"X-Real-Ip",
		"proxy-client-ip",
		"PROXY-CLIENT-IP",
		"Proxy-Client-Ip"} {
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
