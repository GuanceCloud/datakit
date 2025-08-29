// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

type option struct {
	source    string
	timeout   time.Duration
	keepAlive time.Duration

	ignoreReqErr             bool
	metricTypes              []string
	metricNameReFilter       []*regexp.Regexp
	metricNameReFilterIgnore []*regexp.Regexp
	measurementPrefix        string
	measurementName          string
	measurements             []Rule
	keepExistMetricName      bool
	honorTimestamps          bool
	output                   string
	maxFileSize              int64

	udsPath            string
	tlsOpen            bool
	cacertFiles        []string
	certFile           string
	keyFile            string
	insecureSkipVerify bool
	tlsClientConfig    *dknet.TLSClientConfig

	tagsIgnore  []string // do not keep these tags in scraped prom data
	tagsRename  *RenameTags
	asLogging   *AsLogging
	ignoreTagKV map[string][]*regexp.Regexp // drop scraped prom data if tag key's value matched
	httpHeaders map[string]string

	tags           map[string]string
	disableInfoTag bool

	auth map[string]string

	batchCallback func([]*point.Point) error
	streamSize    int
	ptts          int64

	l *logger.Logger
}

type PromOption func(opt *option)

var minimumHTTPTimeout = time.Second * 3

func defaultOption() *option {
	return &option{
		l:       logger.DefaultSLogger("prom"),
		timeout: minimumHTTPTimeout,
	}
}

func WithSource(str string) PromOption { return func(opt *option) { opt.source = str } }
func WithTimeout(dur time.Duration) PromOption {
	return func(opt *option) {
		if minimumHTTPTimeout < dur {
			opt.timeout = dur
		}
	}
}

// WithTimestamp set point's timestamp(nano-seconds) on each scrap.
func WithTimestamp(ts int64) PromOption {
	return func(opt *option) {
		if ts > 0 {
			opt.ptts = ts
		}
	}
}

func WithKeepAlive(dur time.Duration) PromOption {
	return func(opt *option) {
		if dur > 0 {
			opt.keepAlive = dur
		}
	}
}
func WithIgnoreReqErr(b bool) PromOption       { return func(opt *option) { opt.ignoreReqErr = b } }
func WithMetricTypes(strs []string) PromOption { return func(opt *option) { opt.metricTypes = strs } }
func WithMetricNameFilter(strs []string) PromOption {
	return func(opt *option) {
		if len(strs) == 0 {
			return
		}

		opt.metricNameReFilter = make([]*regexp.Regexp, 0, len(strs))
		for _, x := range strs {
			if re, err := regexp.Compile(x); err != nil {
				if opt.l != nil {
					opt.l.Warnf("regexp.Compile('%s'): %s, ignored", x, err)
				}
			} else {
				opt.metricNameReFilter = append(opt.metricNameReFilter, re)
			}
		}
	}
}

func WithMetricNameFilterIgnore(strs []string) PromOption {
	return func(opt *option) {
		if len(strs) == 0 {
			return
		}

		opt.metricNameReFilterIgnore = make([]*regexp.Regexp, 0, len(strs))
		for _, x := range strs {
			if re, err := regexp.Compile(x); err != nil {
				if opt.l != nil {
					opt.l.Warnf("regexp.Compile('%s'): %s, ignored", x, err)
				}
			} else {
				opt.metricNameReFilterIgnore = append(opt.metricNameReFilterIgnore, re)
			}
		}
	}
}

func WithMeasurementPrefix(str string) PromOption {
	return func(opt *option) { opt.measurementPrefix = str }
}

func WithMeasurementName(str string) PromOption {
	return func(opt *option) { opt.measurementName = str }
}

func KeepExistMetricName(b bool) PromOption {
	return func(opt *option) { opt.keepExistMetricName = b }
}

func HonorTimestamps(b bool) PromOption {
	return func(opt *option) { opt.honorTimestamps = b }
}

func WithMeasurements(r []Rule) PromOption    { return func(opt *option) { opt.measurements = r } }
func WithOutput(str string) PromOption        { return func(opt *option) { opt.output = str } }
func WithMaxFileSize(i int64) PromOption      { return func(opt *option) { opt.maxFileSize = i } }
func WithTLSOpen(b bool) PromOption           { return func(opt *option) { opt.tlsOpen = b } }
func WithUDSPath(str string) PromOption       { return func(opt *option) { opt.udsPath = str } }
func WithCacertFiles(arr []string) PromOption { return func(opt *option) { opt.cacertFiles = arr } }
func WithCertFile(str string) PromOption      { return func(opt *option) { opt.certFile = str } }
func WithKeyFile(str string) PromOption       { return func(opt *option) { opt.keyFile = str } }
func WithTLSClientConfig(t *dknet.TLSClientConfig) PromOption {
	return func(opt *option) { opt.tlsClientConfig = t }
}

func WithInsecureSkipVerify(b bool) PromOption {
	return func(opt *option) { opt.insecureSkipVerify = b }
}

func WithBearerToken(str string) PromOption {
	return func(opt *option) {
		if opt.httpHeaders == nil {
			opt.httpHeaders = make(map[string]string)
		}
		opt.httpHeaders["Authorization"] = "Bearer " + str
	}
}

func WithTagsIgnore(strs []string) PromOption { return func(opt *option) { opt.tagsIgnore = strs } }
func WithTagsRename(renameTags *RenameTags) PromOption {
	return func(opt *option) { opt.tagsRename = renameTags }
}

func WithAsLogging(asLogging *AsLogging) PromOption {
	return func(opt *option) { opt.asLogging = asLogging }
}

func WithIgnoreTagKV(tagKVs map[string][]string) PromOption {
	return func(opt *option) {
		kvIgnore := IgnoreTagKeyValMatch{}
		for k, arr := range tagKVs {
			for _, x := range arr {
				if re, err := regexp.Compile(x); err != nil {
					if opt.l != nil {
						opt.l.Warnf("regexp.Compile('%s'): %s, ignored", x, err)
					}
				} else {
					kvIgnore[k] = append(kvIgnore[k], re)
				}
			}
		}
		opt.ignoreTagKV = kvIgnore
	}
}

func WithHTTPHeaders(m map[string]string) PromOption {
	return func(opt *option) { opt.httpHeaders = m }
}
func WithTags(m map[string]string) PromOption { return func(opt *option) { opt.tags = m } }
func WithDisableInfoTag(b bool) PromOption    { return func(opt *option) { opt.disableInfoTag = b } }
func WithAuth(m map[string]string) PromOption { return func(opt *option) { opt.auth = m } }
func WithMaxBatchCallback(i int, f func([]*point.Point) error) PromOption {
	return func(opt *option) {
		if i > 0 && f != nil {
			opt.streamSize = i
			opt.batchCallback = f
		}
	}
}
func WithLogger(l *logger.Logger) PromOption { return func(opt *option) { opt.l = l } }
