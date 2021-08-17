package http

import (
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	// HttpProxy    string
	// proxyOnce    sync.Once
	// proxyFunc    func(*http.Request) (*url.URL, error)
	DefTransport = &http.Transport{
		// Proxy: getProxyURL(),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   45,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
)

// func getProxyURL() func(req *http.Request) (*url.URL, error) {
// 	proxyOnce.Do(func() {
// 		if HttpProxy != "" {
// 			if pxurl, err := url.ParseRequestURI(HttpProxy); err == nil {
// 				proxyFunc = func(*http.Request) (*url.URL, error) {
// 					return pxurl, nil
// 				}
// 			}
// 		}
// 	})

// 	return proxyFunc
// }

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
	for _, h := range []string{"x-forwarded-for", "X-FORWARDED-FOR", "X-Forwarded-For", "x-real-ip", "X-REAL-IP", "X-Real-Ip", "proxy-client-ip", "PROXY-CLIENT-IP", "Proxy-Client-Ip"} {
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
