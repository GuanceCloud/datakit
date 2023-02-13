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

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var dkioFeed func(name, category string, pts []*point.Point, opt *dkio.Option) error = dkio.Feed

type AfterGatherHandler interface {
	Run(inputName string, dktraces DatakitTraces, strikMod bool)
}

type AfterGatherFunc func(inputName string, dktraces DatakitTraces, strikMod bool)

func (ag AfterGatherFunc) Run(inputName string, dktraces DatakitTraces, strikMod bool) {
	ag(inputName, dktraces, strikMod)
}

type AfterGather struct {
	sync.Mutex
	log            *logger.Logger
	filters        []FilterFunc
	ReFeedInterval time.Duration
	BlockIOModel   bool
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

func (aga *AfterGather) Run(inputName string, dktraces DatakitTraces, stricktMod bool) {
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

	if pts := aga.BuildPointsBatch(afterFilters, stricktMod); len(pts) != 0 {
		var (
			start = time.Now()
			opt   = &dkio.Option{Blocking: aga.BlockIOModel}
			err   error
		)
	IO_FEED_RETRY:
		if err = dkioFeed(inputName, datakit.Tracing, pts, opt); err != nil {
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
			if pt, err := BuildPoint(dktraces[i][j], strict); err != nil {
				aga.log.Warnf("build point error: %s", err.Error())
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
	// exclude span_type in tags, span_type is crucial in data display
	if dkspan.SpanType == "" {
		dkspan.SpanType = SPAN_TYPE_UNKNOW
	}
	tags[TAG_SPAN_TYPE] = dkspan.SpanType

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
