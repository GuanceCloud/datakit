// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

type option struct {
	optionClientConn

	source, remote      string
	measurement         string
	keepExistMetricName bool

	extraTags map[string]string
	callback  func([]*point.Point) error
}

type optionClientConn struct {
	timeout   time.Duration
	keepAlive time.Duration

	tlsOpen            bool
	cacertFiles        []string
	certFile           string
	keyFile            string
	insecureSkipVerify bool
	tlsClientConfig    *dknet.TLSClientConfig

	httpHeaders map[string]string
}

type Option func(opt *option)

var discardPointsFn = func([]*point.Point) error {
	return fmt.Errorf("discard points")
}

func defaultOption() *option {
	return &option{
		source: "promscrape",
		optionClientConn: optionClientConn{
			timeout:     time.Second * 10,
			httpHeaders: make(map[string]string),
		},
		extraTags: make(map[string]string),
		callback:  discardPointsFn,
	}
}

func WithSource(str string) Option      { return func(opt *option) { opt.source = str } }
func WithRemote(str string) Option      { return func(opt *option) { opt.remote = str } }
func WithMeasurement(str string) Option { return func(opt *option) { opt.measurement = str } }
func KeepExistMetricName(b bool) Option {
	return func(opt *option) { opt.keepExistMetricName = b }
}

func WithTimeout(dur time.Duration) Option {
	return func(opt *option) {
		if dur > 0 {
			opt.timeout = dur
		}
	}
}

func WithKeepAlive(dur time.Duration) Option {
	return func(opt *option) {
		if dur > 0 {
			opt.keepAlive = dur
		}
	}
}

func WithTLSOpen(b bool) Option           { return func(opt *option) { opt.tlsOpen = b } }
func WithCacertFiles(arr []string) Option { return func(opt *option) { opt.cacertFiles = arr } }
func WithCertFile(str string) Option      { return func(opt *option) { opt.certFile = str } }
func WithKeyFile(str string) Option       { return func(opt *option) { opt.keyFile = str } }

func WithTLSClientConfig(t *dknet.TLSClientConfig) Option {
	return func(opt *option) { opt.tlsClientConfig = t }
}

func WithInsecureSkipVerify(b bool) Option {
	return func(opt *option) { opt.insecureSkipVerify = b }
}

func WithHTTPHeader(m map[string]string) Option {
	return func(opt *option) {
		for k, v := range m {
			opt.httpHeaders[k] = v
		}
	}
}

func WithBearerToken(str string) Option {
	return func(opt *option) {
		opt.httpHeaders["Authorization"] = "Bearer " + str
	}
}

func WithExtraTags(m map[string]string) Option {
	return func(opt *option) {
		for k, v := range m {
			opt.extraTags[k] = v
		}
	}
}

func WithCallback(fn func([]*point.Point) error) Option {
	return func(opt *option) {
		if fn != nil {
			opt.callback = fn
		}
	}
}
