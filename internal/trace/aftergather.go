// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var (
	once                                                                             = sync.Once{}
	dkioFeed func(name, category string, pts []*point.Point, opt *dkio.Option) error = dkio.Feed
)

type AfterGatherHandler interface {
	Run(inputName string, dktraces DatakitTraces, strikMod bool)
}

type AfterGatherFunc func(inputName string, dktraces DatakitTraces, strikMod bool)

func (ag AfterGatherFunc) Run(inputName string, dktraces DatakitTraces, strikMod bool) {
	ag(inputName, dktraces, strikMod)
}

// CalculatorFunc is func type for calculation, statistics, etc
// any data changes in DatakitTraces will be saved and affect the next actions afterwards.
type CalculatorFunc func(dktrace DatakitTrace)

// FilterFunc is func type for data filter.
// Return the DatakitTraces that need to propagate to next action and
// return ture if one want to skip all FilterFunc afterwards, false otherwise.
type FilterFunc func(dktrace DatakitTrace) (DatakitTrace, bool)

type AfterGather struct {
	sync.Mutex
	calculators    []CalculatorFunc
	filters        []FilterFunc
	ReFeedInterval time.Duration
}

type Option func(aga *AfterGather)

func WithRetry(interval time.Duration) Option {
	return func(aga *AfterGather) {
		aga.ReFeedInterval = interval
	}
}

func NewAfterGather(options ...Option) *AfterGather {
	aga := &AfterGather{}
	for i := range options {
		options[i](aga)
	}

	return aga
}

// AppendCalculator will append new calculators into AfterGather structure,
// and run them as the order they added.%.
func (aga *AfterGather) AppendCalculator(calc ...CalculatorFunc) {
	aga.Lock()
	defer aga.Unlock()

	aga.calculators = append(aga.calculators, calc...)
}

// AppendFilter will append new filters into AfterGather structure
// and run them as the order they added. If one filter func return false then
// the filters loop will break.
func (aga *AfterGather) AppendFilter(filter ...FilterFunc) {
	aga.Lock()
	defer aga.Unlock()

	aga.filters = append(aga.filters, filter...)
}

func (aga *AfterGather) Run(inputName string, dktraces DatakitTraces, stricktMod bool) {
	once.Do(func() {
		log = logger.SLogger(packageName)
	})

	if len(dktraces) == 0 {
		log.Debug("empty dktraces")

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
				if temp, skip = aga.filters[i](dktraces[k]); skip {
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

	if pts := BuildPointsBatch(afterFilters, stricktMod); len(pts) != 0 {
		var (
			start = time.Now()
			err   error
		)
	IO_FEED_RETRY:
		if err = dkioFeed(inputName, datakit.Tracing, pts, nil); err != nil {
			log.Warnf("### io feed points failed: %s, ignored", err.Error())
			if aga.ReFeedInterval > 0 && errors.Is(err, dkio.ErrIOBusy) {
				time.Sleep(aga.ReFeedInterval)
				goto IO_FEED_RETRY
			}
		} else {
			log.Debugf("### send %d points cost %dms with error: %v", len(pts), time.Since(start)/time.Millisecond, err)
		}
	} else {
		log.Debug("### BuildPointsBatch return empty points array")
	}
}

// BuildPointsBatch builds points from whole trace.
func BuildPointsBatch(dktraces DatakitTraces, strict bool) []*point.Point {
	var pts []*point.Point
	for i := range dktraces {
		for j := range dktraces[i] {
			if pt, err := BuildPoint(dktraces[i][j], strict); err != nil {
				log.Warnf("build point error: %s", err.Error())
			} else {
				pts = append(pts, pt)
			}
		}
	}

	return pts
}

// BuildPoint builds point from DatakitSpan.
func BuildPoint(dkspan *DatakitSpan, strict bool) (*point.Point, error) {
	if dkspan.Service == "" {
		dkspan.Service = UnknowServiceName(dkspan)
	}

	tags := map[string]string{
		TAG_SERVICE:     dkspan.Service,
		TAG_OPERATION:   dkspan.Operation,
		TAG_SOURCE_TYPE: dkspan.SourceType,
		TAG_SPAN_STATUS: dkspan.Status,
		TAG_SPAN_TYPE:   dkspan.SpanType,
	}
	if dkspan.SpanType == "" {
		tags[TAG_SPAN_TYPE] = SPAN_TYPE_UNKNOW
	}
	if dkspan.SourceType == "" {
		tags[TAG_SOURCE_TYPE] = SPAN_SOURCE_CUSTOMER
	}
	for k, v := range dkspan.Tags {
		if strings.Contains(k, ".") {
			tags[strings.ReplaceAll(k, ".", "_")] = v
		} else {
			tags[k] = v
		}
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
	for k, v := range dkspan.Metrics {
		fields[k] = v
	}

	return point.NewPoint(dkspan.Source, tags, fields, &point.PointOption{
		Time:     time.Unix(0, dkspan.Start),
		Category: datakit.Tracing,
		Strict:   strict,
	})
}
