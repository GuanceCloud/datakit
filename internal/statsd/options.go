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
	protocol               string
	serviceAddress         string
	statsdSourceKey        string
	statsdHostKey          string
	saveAboveKey           bool
	allowedPendingMessages int
	percentiles            []float64
	percentileLimit        int
	deleteGauges           bool
	deleteCounters         bool
	setCounterInt          bool
	deleteSets             bool
	deleteTimings          bool
	convertNames           bool
	metricSeparator        string
	dataDogExtensions      bool
	dataDogDistributions   bool
	udpPacketSize          int
	readBufferSize         int
	dropTags               []string
	metricMapping          []string
	tags                   map[string]string
	maxTCPConnections      int
	tcpKeepAlive           bool
	maxTTL                 time.Duration

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

func WithSetCounterInt(args bool) CollectorOption {
	return func(opt *option) { opt.setCounterInt = args }
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

func WithMaxTTL(args time.Duration) CollectorOption {
	return func(opt *option) {
		opt.maxTTL = args
	}
}

func WithLogger(args *logger.Logger) CollectorOption {
	return func(opt *option) { opt.l = args }
}
