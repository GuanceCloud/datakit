// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace for DK trace.
package trace

import (
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type AfterGatherHandler interface {
	Run(inputName string, dktraces DatakitTraces)
}

type AfterGatherFunc func(inputName string, dktraces DatakitTraces)

func (ag AfterGatherFunc) Run(inputName string, dktraces DatakitTraces) {
	ag(inputName, dktraces)
}

type Option func(aga *AfterGather)

func WithLogger(log *logger.Logger) Option {
	return func(aga *AfterGather) {
		aga.log = log
	}
}

func WithRetry(interval time.Duration) Option {
	return func(aga *AfterGather) {
		aga.retry = interval
	}
}

func WithPointOptions(opts ...point.Option) Option {
	return func(aga *AfterGather) {
		aga.pointOptions = append(aga.pointOptions, opts...)
	}
}

func WithFeeder(feeder dkio.Feeder) Option {
	return func(aga *AfterGather) {
		aga.feeder = feeder
	}
}

type AfterGather struct {
	sync.Mutex
	log          *logger.Logger
	filters      []FilterFunc
	retry        time.Duration
	pointOptions []point.Option
	feeder       dkio.Feeder
}

// AppendFilter will append new filters into AfterGather structure
// and run them as the order they added. If one filter func return false then
// the filters loop will break.
func (aga *AfterGather) AppendFilter(filter ...FilterFunc) {
	aga.Lock()
	defer aga.Unlock()

	aga.filters = append(aga.filters, filter...)
}

func (aga *AfterGather) doFeed(iname string, dktrace DatakitTrace) {
	var pts []*point.Point
	for _, span := range dktrace {
		span.Point.AddTag(TagDKFingerprintKey, datakit.DKHost)
		pts = append(pts, span.Point)
	}

	if err := aga.feeder.Feed(point.Tracing, pts, dkio.WithSource(iname)); err != nil {
		aga.log.Warnf("feed %d points failed: %s, ignored", len(pts), err.Error())
	}
}

func (aga *AfterGather) Run(inputName string, dktraces DatakitTraces) {
	if len(dktraces) == 0 {
		aga.log.Debug("empty dktraces")

		return
	}

	var afterFilters DatakitTraces
	if len(aga.filters) == 0 {
		afterFilters = dktraces
	} else {
		afterFilters = make(DatakitTraces, 0, len(dktraces))

		for k := range dktraces {
			aga.log.Debugf("len = %d spans", len(dktraces[k]))
			serviceName := dktraces[k][0].GetTag(TagService)

			TracingProcessCount.WithLabelValues(inputName, serviceName).Add(1)

			var singleTrace DatakitTrace
			for i := range aga.filters {
				var skip bool
				if singleTrace, skip = aga.filters[i](aga.log, dktraces[k]); skip {
					break // skip current trace
				}
			}

			if singleTrace != nil {
				afterFilters = append(afterFilters, singleTrace)
			}
		}
	}

	for _, trace := range afterFilters {
		aga.doFeed(inputName, trace)
	}
}

func NewAfterGather(options ...Option) *AfterGather {
	aga := &AfterGather{log: logger.DefaultSLogger("after-gather")}
	for i := range options {
		options[i](aga)
	}

	return aga
}

var replacer = strings.NewReplacer(".", "_")
