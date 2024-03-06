// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace for DK trace.
package trace

import (
	"errors"
	"strings"
	"sync"
	"time"

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

func WithIOBlockingMode(block bool) Option {
	return func(aga *AfterGather) {
		aga.ioBlockingMode = block
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
	log            *logger.Logger
	filters        []FilterFunc
	retry          time.Duration
	ioBlockingMode bool
	pointOptions   []point.Option
	feeder         dkio.Feeder
}

// AppendFilter will append new filters into AfterGather structure
// and run them as the order they added. If one filter func return false then
// the filters loop will break.
func (aga *AfterGather) AppendFilter(filter ...FilterFunc) {
	aga.Lock()
	defer aga.Unlock()

	aga.filters = append(aga.filters, filter...)
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
		for k := range dktraces {
			aga.log.Debugf("len = %d spans", len(dktraces[k]))
			var temp DatakitTrace
			for i := range aga.filters {
				var skip bool
				if temp, skip = aga.filters[i](aga.log, dktraces[k]); skip {
					break
				}
			}
			if temp != nil {
				afterFilters = append(afterFilters, temp)
			}
		}
	}
	if len(afterFilters) == 0 {
		return
	}
	pts := make([]*point.Point, 0)
	for _, filter := range afterFilters {
		for _, span := range filter {
			pts = append(pts, span.Point)
		}
	}
	if len(pts) != 0 {
		var (
			start = time.Now()
			err   error
		)
	IO_FEED_RETRY:
		if err = aga.feeder.FeedV2(point.Tracing, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithBlocking(aga.ioBlockingMode),
			dkio.WithInputName(inputName)); err != nil {
			if aga.retry > 0 && errors.Is(err, dkio.ErrIOBusy) {
				time.Sleep(aga.retry)
				goto IO_FEED_RETRY
			}
		} else {
			aga.log.Debugf("### send %d points cost %dms", len(pts), time.Since(start)/time.Millisecond)
		}
	} else {
		aga.log.Debug("BuildPointsBatch return empty points array")
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
