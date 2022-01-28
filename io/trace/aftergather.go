package trace

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

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
	if inputName == "" || len(dktraces) == 0 {
		return
	}

	for i := range ag.Calculators {
		ag.Calculators[i](dktraces)
	}
	for i := range ag.Filters {
		dktraces = ag.Filters[i](dktraces)
	}

	MakeLineProto(inputName, dktraces)
}

func BuildLineProto(dkspan *DatakitSpan) (*dkio.Point, error) {
	dkOnce.Do(func() {
		log = logger.SLogger("dktrace")
	})

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

	pt, err := dkio.NewPoint(dkspan.Source, tags, fields, &dkio.PointOption{
		Time:     time.Unix(0, dkspan.Start),
		Category: datakit.Tracing,
		Strict:   false,
	})
	if err != nil {
		log.Errorf("build metric err: %s", err)
	}

	return pt, err
}

func MakeLineProto(inputName string, dktraces DatakitTraces) {
	var pts []*dkio.Point
	for i := range dktraces {
		for j := range dktraces[i] {
			if pt, err := BuildLineProto(dktraces[i][j]); err == nil {
				pts = append(pts, pt)
			}
		}
	}

	if err := dkio.Feed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
		log.Errorf("io feed err: %s", err)
	}
}
