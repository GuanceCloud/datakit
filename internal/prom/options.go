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
)

type option struct {
	source    string
	timeout   time.Duration
	keepAlive time.Duration

	ignoreReqErr           bool
	metricTypes            []string
	metricNameFilter       []string
	metricNameFilterIgnore []string
	measurementPrefix      string
	measurementName        string
	measurements           []Rule
	output                 string
	maxFileSize            int64

	tlsOpen    bool
	udsPath    string
	cacertFile string
	certFile   string
	keyFile    string

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
	l             *logger.Logger
}

type PromOption func(opt *option)

func WithSource(str string) PromOption { return func(opt *option) { opt.source = str } }
func WithTimeout(dura time.Duration) PromOption {
	return func(opt *option) {
		if dura > 0 {
			opt.timeout = dura
		}
	}
}

func WithKeepAlive(dura time.Duration) PromOption {
	return func(opt *option) {
		if dura > 0 {
			opt.keepAlive = dura
		}
	}
}
func WithIgnoreReqErr(b bool) PromOption       { return func(opt *option) { opt.ignoreReqErr = b } }
func WithMetricTypes(strs []string) PromOption { return func(opt *option) { opt.metricTypes = strs } }
func WithMetricNameFilter(strs []string) PromOption {
	return func(opt *option) { opt.metricNameFilter = strs }
}

func WithMetricNameFilterIgnore(strs []string) PromOption {
	return func(opt *option) { opt.metricNameFilterIgnore = strs }
}

func WithMeasurementPrefix(str string) PromOption {
	return func(opt *option) { opt.measurementPrefix = str }
}

func WithMeasurementName(str string) PromOption {
	return func(opt *option) { opt.measurementName = str }
}
func WithMeasurements(r []Rule) PromOption    { return func(opt *option) { opt.measurements = r } }
func WithOutput(str string) PromOption        { return func(opt *option) { opt.output = str } }
func WithMaxFileSize(i int64) PromOption      { return func(opt *option) { opt.maxFileSize = i } }
func WithTLSOpen(b bool) PromOption           { return func(opt *option) { opt.tlsOpen = b } }
func WithUDSPath(str string) PromOption       { return func(opt *option) { opt.udsPath = str } }
func WithCacertFile(str string) PromOption    { return func(opt *option) { opt.cacertFile = str } }
func WithCertFile(str string) PromOption      { return func(opt *option) { opt.certFile = str } }
func WithKeyFile(str string) PromOption       { return func(opt *option) { opt.keyFile = str } }
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
