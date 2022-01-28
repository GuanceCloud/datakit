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
	sync.Mutex
	Calculators map[string]CalculatorFunc
	Filters     map[string]FilterFunc
}

func NewAfterGather() *AfterGather {
	return &AfterGather{
		Calculators: make(map[string]CalculatorFunc),
		Filters:     make(map[string]FilterFunc),
	}
}

func (ag *AfterGather) AddCalculator(key string, calc CalculatorFunc) {
	ag.Lock()
	defer ag.Unlock()

	ag.Calculators[key] = calc
}

func (ag *AfterGather) DelCalculator(key string) {
	ag.Lock()
	defer ag.Unlock()

	delete(ag.Calculators, key)
}

func (ag *AfterGather) AddFilter(key string, filter FilterFunc) {
	ag.Lock()
	defer ag.Unlock()

	ag.Filters[key] = filter
}

func (ag *AfterGather) DelFilter(key string) {
	ag.Lock()
	defer ag.Unlock()

	delete(ag.Filters, key)
}

func (ag *AfterGather) Run(inputName string, dktraces DatakitTraces, stricktMod bool) {
	once.Do(func() {
		log = logger.SLogger(packageName)
	})

	if inputName == "" || len(dktraces) == 0 {
		log.Warnf("wrong parameters for AfterGather.Run(inputName: %s, dktraces:%v)", inputName, dktraces)

		return
	}

	for i := range ag.Calculators {
		ag.Calculators[i](dktraces)
	}
	for i := range ag.Filters {
		dktraces = ag.Filters[i](dktraces)
	}

	if pts := BuildPointsBatch(inputName, dktraces, stricktMod); len(pts) != 0 {
		if err := dkio.Feed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true}); err != nil {
			log.Errorf("io feed points error: %s", err.Error())
		}
	} else {
		log.Warn("empty points")
	}
}

func BuildPointsBatch(inputName string, dktraces DatakitTraces, strict bool) []*dkio.Point {
	var pts []*dkio.Point
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

func BuildPoint(dkspan *DatakitSpan, strict bool) (*dkio.Point, error) {
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
		Strict:   strict,
	})
}
