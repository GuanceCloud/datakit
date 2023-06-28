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

////////////////////////////////////////////////////////////////////////////////

type AfterGatherHandler interface {
	Run(inputName string, dktraces DatakitTraces, strictMod bool)
}

type AfterGatherFunc func(inputName string, dktraces DatakitTraces, strictMod bool)

func (ag AfterGatherFunc) Run(inputName string, dktraces DatakitTraces, strictMod bool) {
	ag(inputName, dktraces, strictMod)
}

////////////////////////////////////////////////////////////////////////////////

type AfterGather struct {
	sync.Mutex
	log            *logger.Logger
	filters        []FilterFunc
	ReFeedInterval time.Duration
	BlockIOModel   bool
	inputOption    point.Option
	feeder         dkio.Feeder
}

type Option func(aga *AfterGather)

func WithLogger(log *logger.Logger) Option {
	return func(aga *AfterGather) {
		aga.log = log
	}
}

func WithRetry(interval time.Duration) Option {
	return func(aga *AfterGather) {
		aga.ReFeedInterval = interval
	}
}

func WithBlockIOModel(block bool) Option {
	return func(aga *AfterGather) {
		aga.BlockIOModel = block
	}
}

func WithInputOption(opt point.Option) Option {
	return func(aga *AfterGather) {
		aga.inputOption = opt
	}
}

func WithFeeder(feeder dkio.Feeder) Option {
	return func(aga *AfterGather) {
		aga.feeder = feeder
	}
}

func NewAfterGather(options ...Option) *AfterGather {
	aga := &AfterGather{log: logger.DefaultSLogger("after-gather")}
	for i := range options {
		options[i](aga)
	}

	return aga
}

// AppendFilter will append new filters into AfterGather structure
// and run them as the order they added. If one filter func return false then
// the filters loop will break.
func (aga *AfterGather) AppendFilter(filter ...FilterFunc) {
	aga.Lock()
	defer aga.Unlock()

	aga.filters = append(aga.filters, filter...)
}

func (aga *AfterGather) Run(inputName string, dktraces DatakitTraces, strictMod bool) {
	if len(dktraces) == 0 {
		aga.log.Debug("empty dktraces")

		return
	}

	var afterFilters DatakitTraces
	if len(aga.filters) == 0 {
		afterFilters = dktraces
	} else {
		for k := range dktraces {
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

	if pts := aga.BuildPointsBatch(afterFilters, strictMod); len(pts) != 0 {
		var (
			start = time.Now()
			opt   = &dkio.Option{Blocking: aga.BlockIOModel}
			err   error
		)
	IO_FEED_RETRY:
		if err = aga.feeder.Feed(inputName, point.Tracing, pts, opt); err != nil {
			aga.log.Warnf("io feed points failed: %s, ignored", err.Error())
			if aga.ReFeedInterval > 0 && errors.Is(err, dkio.ErrIOBusy) {
				time.Sleep(aga.ReFeedInterval)
				goto IO_FEED_RETRY
			}
		} else {
			aga.log.Debugf("### send %d points cost %dms with error: %v", len(pts), time.Since(start)/time.Millisecond, err)
		}
	} else {
		aga.log.Debug("BuildPointsBatch return empty points array")
	}
}

// BuildPointsBatch builds points from whole trace.
func (aga *AfterGather) BuildPointsBatch(dktraces DatakitTraces, strict bool) []*point.Point {
	var pts []*point.Point
	for i := range dktraces {
		for j := range dktraces[i] {
			if pt, err := BuildPoint(dktraces[i][j], strict, aga.inputOption); err != nil {
				aga.log.Warnf("build point error: %s", err.Error())
			} else {
				pts = append(pts, pt)
			}
		}
	}

	return pts
}

////////////////////////////////////////////////////////////////////////////////

func processUnknown(dkspan *DatakitSpan) {
	if dkspan != nil {
		if dkspan.Service == "" {
			dkspan.Service = UNKNOWN_SERVICE
		}
		if dkspan.SourceType == "" {
			dkspan.SourceType = SPAN_SOURCE_CUSTOMER
		}
		if dkspan.SpanType == "" {
			dkspan.SpanType = SPAN_TYPE_UNKNOWN
		}
	}
}

// BuildPoint builds point from DatakitSpan.
func BuildPoint(dkspan *DatakitSpan, strict bool, optFromInput point.Option) (*point.Point, error) {
	processUnknown(dkspan)

	tags := map[string]string{
		TAG_SERVICE:     dkspan.Service,
		TAG_OPERATION:   dkspan.Operation,
		TAG_SOURCE_TYPE: dkspan.SourceType,
		TAG_SPAN_TYPE:   dkspan.SpanType,
		TAG_SPAN_STATUS: dkspan.Status,
	}

	for k, v := range dkspan.Tags {
		tags[strings.ReplaceAll(k, ".", "_")] = v
	}

	fields := map[string]interface{}{
		FIELD_TRACEID:  dkspan.TraceID,
		FIELD_PARENTID: dkspan.ParentID,
		FIELD_SPANID:   dkspan.SpanID,
		FIELD_RESOURCE: dkspan.Resource,
		FIELD_START:    dkspan.Start / int64(time.Microsecond),
		FIELD_DURATION: dkspan.Duration / int64(time.Microsecond),
		FIELD_MESSAGE:  dkspan.Content,
	}

	// trace-128-id replace trace-id.
	if id, ok := dkspan.Tags[TRACE_128_BIT_ID]; ok {
		fields[FIELD_TRACEID] = id
	}

	for k, v := range dkspan.Metrics {
		fields[strings.ReplaceAll(k, ".", "_")] = v
	}

	tracing := &TraceMeasurement{
		Name:   dkspan.Source,
		Tags:   tags,
		Fields: fields,
		TS:     time.Unix(0, dkspan.Start),
		Opt:    optFromInput,
	}

	return tracing.Point(), nil
}
