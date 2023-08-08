// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

type option struct {
	// Protocol used on listener - udp or tcp
	protocol string `toml:"protocol"`

	// Address & Port to serve from
	serviceAddress string `toml:"service_address"`

	// Tag request metric. Used for distinguish feed metric name.
	statsdSourceKey string `toml:"statsd_source_key"`
	statsdHostKey   string `toml:"statsd_host_key"`
	saveAboveKey    bool   `toml:"save_above_key"`

	// Number of messages allowed to queue up in between calls to Gather. If this
	// fills up, packets will get dropped until the next Gather interval is ran.
	allowedPendingMessages int `toml:"allowed_pending_messages"`

	// Percentiles specifies the percentiles that will be calculated for timing
	// and histogram stats.
	percentiles     []float64 `toml:"percentiles"`
	percentileLimit int       `toml:"percentile_limit"`
	deleteGauges    bool      `toml:"delete_gauges"`
	deleteCounters  bool      `toml:"delete_counters"`
	deleteSets      bool      `toml:"delete_sets"`
	deleteTimings   bool      `toml:"delete_timings"`
	convertNames    bool

	// MetricSeparator is the separator between parts of the metric name.
	metricSeparator string `toml:"metric_separator"`

	// Parses extensions to statsd in the datadog statsd format
	// currently supports metrics and datadog tags.
	// http://docs.datadoghq.com/guides/dogstatsd/
	dataDogExtensions bool `toml:"datadog_extensions"`

	// Parses distribution metrics in the datadog statsd format.
	// Requires the DataDogExtension flag to be enabled.
	// https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
	dataDogDistributions bool `toml:"datadog_distributions"`

	// UDPPacketSize is deprecated, it's only here for legacy support
	// we now always create 1 max size buffer and then copy only what we need
	// into the in channel
	// see https://github.com/influxdata/telegraf/pull/992
	udpPacketSize int `toml:"udp_packet_size"`

	readBufferSize int `toml:"read_buffer_size"`

	dropTags      []string `toml:"drop_tags"`
	metricMapping []string `toml:"metric_mapping"`

	tags              map[string]string `toml:"tags"`
	maxTCPConnections int               `toml:"max_tcp_connections"`

	tcpKeepAlive bool `toml:"tcp_keep_alive"`
	// tcpKeepAlivePeriod *time.Duration `toml:"tcp_keep_alive_period"`

	// Max duration for each metric to stay cached without being updated.
	maxTTL time.Duration `toml:"max_ttl"`

	l *logger.Logger
}

type CollectorOption func(opt *option)

func WithProtocol(args string) CollectorOption {
	return func(opt *option) { opt.protocol = args }
}

func WithServiceAddress(args string) CollectorOption {
	return func(opt *option) { opt.serviceAddress = args }
}

func WithStatsdSourceKey(args string) CollectorOption {
	return func(opt *option) { opt.statsdSourceKey = args }
}

func WithStatsdHostKey(args string) CollectorOption {
	return func(opt *option) { opt.statsdHostKey = args }
}

func WithSaveAboveKey(args bool) CollectorOption {
	return func(opt *option) { opt.saveAboveKey = args }
}

func WithAllowedPendingMessages(args int) CollectorOption {
	return func(opt *option) { opt.allowedPendingMessages = args }
}

func WithPercentiles(args []float64) CollectorOption {
	return func(opt *option) { opt.percentiles = args }
}

func WithPercentileLimit(args int) CollectorOption {
	return func(opt *option) { opt.percentileLimit = args }
}

func WithDeleteGauges(args bool) CollectorOption {
	return func(opt *option) { opt.deleteGauges = args }
}

func WithDeleteCounters(args bool) CollectorOption {
	return func(opt *option) { opt.deleteCounters = args }
}

func WithDeleteSets(args bool) CollectorOption {
	return func(opt *option) { opt.deleteSets = args }
}

func WithDeleteTimings(args bool) CollectorOption {
	return func(opt *option) { opt.deleteTimings = args }
}

func WithConvertNames(args bool) CollectorOption {
	return func(opt *option) { opt.convertNames = args }
}

func WithMetricSeparator(args string) CollectorOption {
	return func(opt *option) { opt.metricSeparator = args }
}

func WithDataDogExtensions(args bool) CollectorOption {
	return func(opt *option) { opt.dataDogExtensions = args }
}

func WithDataDogDistributions(args bool) CollectorOption {
	return func(opt *option) { opt.dataDogDistributions = args }
}

func WithUDPPacketSize(args int) CollectorOption {
	return func(opt *option) { opt.udpPacketSize = args }
}

func WithReadBufferSize(args int) CollectorOption {
	return func(opt *option) { opt.readBufferSize = args }
}

func WithDropTags(args []string) CollectorOption {
	return func(opt *option) { opt.dropTags = args }
}

func WithMetricMapping(args []string) CollectorOption {
	return func(opt *option) { opt.metricMapping = args }
}

func WithTags(args map[string]string) CollectorOption {
	return func(opt *option) { opt.tags = args }
}

func WithMaxTCPConnections(args int) CollectorOption {
	return func(opt *option) { opt.maxTCPConnections = args }
}

func WithTCPKeepAlive(args bool) CollectorOption {
	return func(opt *option) { opt.tcpKeepAlive = args }
}

// func WithTCPKeepAlivePeriod(args *time.Duration) CollectorOption {
// 	return func(opt *option) {
// 			opt.tcpKeepAlivePeriod = args
// 	}
// }

func WithMaxTTL(args time.Duration) CollectorOption {
	return func(opt *option) {
		opt.maxTTL = args
	}
}

func WithLogger(args *logger.Logger) CollectorOption {
	return func(opt *option) { opt.l = args }
}
