// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpcli

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

const DefaultKeepAlive = 90 * time.Second

type Options struct {
	DialTimeout   time.Duration
	DialKeepAlive time.Duration

	MaxIdleConns        int
	MaxIdleConnsPerHost int

	MockedDelay,
	IdleConnTimeout,
	TLSHandshakeTimeout,
	ExpectContinueTimeout time.Duration

	ProxyURL        *url.URL
	DialContext     dnet.DialFunc
	TLSClientConfig *tls.Config

	NullTransport bool
	Logger        *logger.Logger
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
		tlsConfig   *tls.Config
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

	if opt.TLSClientConfig != nil {
		tlsConfig = opt.TLSClientConfig.Clone()
	}

	return &http.Transport{
		Proxy:           proxy,
		DialContext:     dialContext,
		TLSClientConfig: tlsConfig,

		MaxIdleConns: func() int {
			if opt.MaxIdleConns == 0 {
				return 100
			}
			return opt.MaxIdleConns
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
	if opt != nil && opt.NullTransport { // use null transport
		t := &nullTransport{}

		if opt.Logger != nil {
			t.l = opt.Logger
		}

		if opt.MockedDelay > 0 {
			t.delay = opt.MockedDelay
		}

		return &http.Client{
			Transport: t,
		}
	} else {
		return &http.Client{
			Transport: Transport(opt),
		}
	}
}

func Transport(opt *Options) *http.Transport {
	if opt == nil {
		return DefTransport()
	}
	return newCliTransport(opt)
}

type nullTransport struct {
	l     *logger.Logger
	delay time.Duration
}

func (t *nullTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body != nil {
		_, err := io.Copy(io.Discard, r.Body) // we do not need the body
		if err != nil {
			return nil, err
		}
	}

	if t.l != nil {
		t.l.Debugf("null transport send request %s | %q ok", r.Method, r.URL.Path)
	}

	if t.delay > 0 {
		now := time.Duration(time.Now().UnixNano())
		time.Sleep(now % t.delay)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Request:    r,
	}, nil
}
