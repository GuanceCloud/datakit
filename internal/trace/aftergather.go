// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
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
	calculators []CalculatorFunc
	filters     []FilterFunc
}

func NewAfterGather() *AfterGather {
	return &AfterGather{}
}

// AppendCalculator will append new calculators into AfterGather structure,
// and run them as the order they added.
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
		log.Warnf("wrong parameters for AfterGather.Run(dktrace:%v)", dktraces)

		return
	}

	// for i := range aga.calculators {
	// 	for k := range dktraces {
	// 		aga.calculators[i](dktraces[k])
	// 	}
	// }

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
		if err := dkioFeed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
			log.Errorf("io feed points error: %s", err.Error())
		}
	} else {
		log.Warn("BuildPointsBatch return empty points array")
	}
}

// BuildPointsBatch builds points from whole trace.
func BuildPointsBatch(dktraces DatakitTraces, strict bool) []*point.Point {
	var pts []*point.Point
	for i := range dktraces {
		for j := range dktraces[i] {
			if pt, err := BuildPoint(dktraces[i][j], strict); err != nil {
				log.Errorf("build point error: %s", err.Error())
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
		TAG_CONTAINER_HOST: dkspan.ContainerHost,
		TAG_ENDPOINT:       dkspan.EndPoint,
		TAG_ENV:            dkspan.Env,
		TAG_HTTP_CODE:      dkspan.HTTPStatusCode,
		TAG_HTTP_METHOD:    dkspan.HTTPMethod,
		TAG_OPERATION:      dkspan.Operation,
		TAG_PROJECT:        dkspan.Project,
		TAG_SERVICE:        dkspan.Service,
		TAG_SOURCE_TYPE:    dkspan.SourceType,
		TAG_SPAN_STATUS:    dkspan.Status,
		TAG_SPAN_TYPE:      dkspan.SpanType,
		TAG_VERSION:        dkspan.Version,
	}
	if dkspan.EndPoint == "" {
		tags[TAG_ENDPOINT] = "null"
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
		FIELD_DURATION: dkspan.Duration / int64(time.Microsecond),
		FIELD_MSG:      dkspan.Content,
		FIELD_PARENTID: dkspan.ParentID,
		FIELD_PID:      dkspan.PID,
		FIELD_RESOURCE: dkspan.Resource,
		FIELD_SPANID:   dkspan.SpanID,
		FIELD_START:    dkspan.Start / int64(time.Microsecond),
		FIELD_TRACEID:  dkspan.TraceID,
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
