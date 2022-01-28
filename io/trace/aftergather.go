package trace

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var once = sync.Once{}

type CalculatorFunc func(dktraces DatakitTraces)

type FilterFunc func(dktraces DatakitTraces) DatakitTraces

type AfterGather struct {
	Calculators []CalculatorFunc
	Filters     []FilterFunc
}

func (ag *AfterGather) AddCalculator(calcs ...CalculatorFunc) {
	ag.Calculators = append(ag.Calculators, calcs...)
}

func (ag *AfterGather) AddFilter(filters ...FilterFunc) {
	ag.Filters = append(ag.Filters, filters...)
}

func (ag *AfterGather) Run(inputName string, dktraces DatakitTraces) {
	once.Do(func() {
		log = logger.SLogger(packageName)
	})

	if inputName == "" || len(dktraces) == 0 {
		return
	}

	for i := range ag.Calculators {
		ag.Calculators[i](dktraces)
	}
	for i := range ag.Filters {
		dktraces = ag.Filters[i](dktraces)
	}

	pts := BuildPointsBatch(inputName, dktraces)
	if len(pts) != 0 {
		if err := dkio.Feed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
			log.Errorf("io feed points error: %s", err.Error())
		}
	} else {
		log.Warn("empty points")
	}
}

func BuildPointsBatch(inputName string, dktraces DatakitTraces) []*dkio.Point {
	var pts []*dkio.Point
	for i := range dktraces {
		for j := range dktraces[i] {
			if pt, err := BuildPoint(dktraces[i][j]); err != nil {
				log.Errorf("build point error: %s", err.Error())
			} else {
				pts = append(pts, pt)
			}
		}
	}

	return pts
}

func BuildPoint(dkspan *DatakitSpan) (*dkio.Point, error) {
	var (
		tags   = make(map[string]string)
		fields = make(map[string]interface{})
	)

	tags[TAG_PROJECT] = dkspan.Project
	tags[TAG_OPERATION] = dkspan.Operation
	tags[TAG_SERVICE] = dkspan.Service
	tags[TAG_VERSION] = dkspan.Version
	tags[TAG_ENV] = dkspan.Env
	tags[TAG_HTTP_METHOD] = dkspan.HTTPMethod
	tags[TAG_HTTP_CODE] = dkspan.HTTPStatusCode

	if dkspan.SourceType != "" {
		tags[TAG_TYPE] = dkspan.SourceType
	} else {
		tags[TAG_TYPE] = SPAN_SERVICE_CUSTOM
	}

	for k, v := range dkspan.Tags {
		tags[k] = v
	}

	tags[TAG_SPAN_STATUS] = dkspan.Status

	if dkspan.EndPoint != "" {
		tags[TAG_ENDPOINT] = dkspan.EndPoint
	} else {
		tags[TAG_ENDPOINT] = "null"
	}

	if dkspan.SpanType != "" {
		tags[TAG_SPAN_TYPE] = dkspan.SpanType
	} else {
		tags[TAG_SPAN_TYPE] = SPAN_TYPE_ENTRY
	}

	if dkspan.ContainerHost != "" {
		tags[TAG_CONTAINER_HOST] = dkspan.ContainerHost
	}

	if dkspan.ParentID == "" {
		dkspan.ParentID = "0"
	}

	fields[FIELD_START] = dkspan.Start / int64(time.Microsecond)
	fields[FIELD_DURATION] = dkspan.Duration / int64(time.Microsecond)
	fields[FIELD_MSG] = dkspan.Content
	fields[FIELD_RESOURCE] = dkspan.Resource
	fields[FIELD_PARENTID] = dkspan.ParentID
	fields[FIELD_TRACEID] = dkspan.TraceID
	fields[FIELD_SPANID] = dkspan.SpanID

	return dkio.NewPoint(dkspan.Source, tags, fields, &dkio.PointOption{
		Time:     time.Unix(0, dkspan.Start),
		Category: datakit.Tracing,
		Strict:   false,
	})
}
