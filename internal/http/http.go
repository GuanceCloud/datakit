package http

import (
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

var (
	DefTransport = defaultCliTransport(&Options{
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
	ProxyURL              *url.URL
}

func defaultCliTransport(opt *Options) *http.Transport {

	var proxy func(*http.Request) (*url.URL, error)

	if opt.ProxyURL != nil {
		proxy = http.ProxyURL(opt.ProxyURL)
	}

	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   opt.DialTimeout,
			KeepAlive: opt.DialKeepAlive,
		}).DialContext,

		Proxy: proxy,

		MaxIdleConns: func() int {
			if opt.MaxIdleConns == 0 {
				return 100
			} else {
				return opt.MaxIdleConns
			}
		}(),
		MaxIdleConnsPerHost: func() int {
			if opt.MaxIdleConnsPerHost == 0 {
				return runtime.NumGoroutine()
			} else {
				return opt.MaxIdleConnsPerHost
			}
		}(),

		IdleConnTimeout:       opt.IdleConnTimeout,
		TLSHandshakeTimeout:   opt.TLSHandshakeTimeout,
		ExpectContinueTimeout: opt.ExpectContinueTimeout,
	}
}

func HTTPCli(opt *Options) *http.Client {
	if opt == nil {
		return &http.Client{
			Transport: DefTransport,
		}
	}

	return &http.Client{
		Transport: defaultCliTransport(opt),
	}
}

func SendRequest(req *http.Request) (*http.Response, error) {
	return (&http.Client{Transport: DefTransport}).Do(req)
}

func SendRequestWithTimeout(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{
		Transport: DefTransport,
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
