package ddtrace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ugorji/go/codec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type Span struct {
	Name     string             `json:"name"`
	Service  string             `json:"service"`
	Resource string             `json:"resource"`
	Type     string             `json:"type"`              // protocol associated with the span
	Start    int64              `json:"start"`             // span start time expressed in nanoseconds since epoch
	Duration int64              `json:"duration"`          // duration of the span expressed in nanoseconds
	Meta     map[string]string  `json:"meta,omitempty"`    // arbitrary map of metadata
	Metrics  map[string]float64 `json:"metrics,omitempty"` // arbitrary map of numeric metrics
	SpanID   uint64             `json:"span_id"`           // identifier of this span
	TraceID  uint64             `json:"trace_id"`          // identifier of the root span
	ParentID uint64             `json:"parent_id"`         // identifier of the span's direct parent
	Error    int32              `json:"error"`             // error status of the span; 0 means no errors
	Sampled  bool               `json:"-"`                 // if this span is sampled (and should be kept/recorded) or not
}

var ddtraceSpanType = map[string]string{
	"cache":         trace.SPAN_SERVICE_CACHE,
	"cassandra":     trace.SPAN_SERVICE_DB,
	"elasticsearch": trace.SPAN_SERVICE_DB,
	"grpc":          trace.SPAN_SERVICE_CUSTOM,
	"http":          trace.SPAN_SERVICE_WEB,
	"mongodb":       trace.SPAN_SERVICE_DB,
	"redis":         trace.SPAN_SERVICE_CACHE,
	"sql":           trace.SPAN_SERVICE_DB,
	"template":      trace.SPAN_SERVICE_CUSTOM,
	"test":          trace.SPAN_SERVICE_CUSTOM,
	"web":           trace.SPAN_SERVICE_WEB,
	"worker":        trace.SPAN_SERVICE_CUSTOM,
	"memcached":     trace.SPAN_SERVICE_CACHE,
	"leveldb":       trace.SPAN_SERVICE_DB,
	"dns":           trace.SPAN_SERVICE_CUSTOM,
	"queue":         trace.SPAN_SERVICE_CUSTOM,
	"consul":        trace.SPAN_SERVICE_APP,
	"rpc":           trace.SPAN_SERVICE_CUSTOM,
	"soap":          trace.SPAN_SERVICE_CUSTOM,
	"db":            trace.SPAN_SERVICE_DB,
	"hibernate":     trace.SPAN_SERVICE_CUSTOM,
	"aerospike":     trace.SPAN_SERVICE_DB,
	"datanucleus":   trace.SPAN_SERVICE_CUSTOM,
	"graphql":       trace.SPAN_SERVICE_CUSTOM,
	"custom":        trace.SPAN_SERVICE_CUSTOM,
	"benchmark":     trace.SPAN_SERVICE_CUSTOM,
	"build":         trace.SPAN_SERVICE_CUSTOM,
	"":              trace.SPAN_SERVICE_CUSTOM,
}

func DdtraceTraceHandle(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleDdtrace(w, r); err != nil {
		dkio.FeedLastError(inputName, err.Error())
		log.Errorf("%v", err)
	}
}

func handleDdtrace(w http.ResponseWriter, r *http.Request) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/msgpack" {
		return parseDdtraceMsgpack(r.Body)
	} else {
		return fmt.Errorf("ddtrace unsupported Content-Type: %s", contentType)
	}
}

func parseDdtraceMsgpack(body io.ReadCloser) error {
	dspans, err := unmarshalDdtraceMsgpack(body)
	if err != nil {
		return err
	}

	pts := []*dkio.Point{}
	for _, spans := range dspans {
		spanIds, parentIds := getSpanAndParentId(spans)
		for _, span := range spans {
			tags := make(map[string]string)
			field := make(map[string]interface{})

			tm := &trace.TraceMeasurement{}
			tm.Name = "ddtrace"

			spanType := ""
			if span.ParentID == 0 {
				spanType = trace.SPAN_TYPE_ENTRY
			} else {
				if serviceName, ok := spanIds[span.ParentID]; ok {
					if serviceName != span.Service {
						spanType = trace.SPAN_TYPE_ENTRY
					} else {
						if _, ok := parentIds[span.SpanID]; ok {
							spanType = trace.SPAN_TYPE_LOCAL
						} else {
							spanType = trace.SPAN_TYPE_EXIT
						}
					}
				} else {
					spanType = trace.SPAN_TYPE_ENTRY
				}
			}
			tags[trace.TAG_SPAN_TYPE] = spanType
			tags[trace.TAG_SERVICE] = span.Service
			tags[trace.TAG_OPERATION] = span.Name
			tags[trace.TAG_TYPE] = ddtraceSpanType[span.Type]
			if span.Error == 0 {
				tags[trace.TAG_SPAN_STATUS] = trace.STATUS_OK
			} else {
				tags[trace.TAG_SPAN_STATUS] = trace.STATUS_ERR
			}
			tags[trace.TAG_PROJECT] = span.Meta[trace.PROJECT]
			if tags[trace.TAG_PROJECT] == "" {
				tags[trace.TAG_PROJECT] = trace.GetFromPluginTag(DdtraceTags, trace.PROJECT)
			}

			tags[trace.TAG_ENV] = span.Meta[trace.ENV]
			if tags[trace.TAG_ENV] == "" {
				tags[trace.TAG_ENV] = trace.GetFromPluginTag(DdtraceTags, trace.ENV)
			}

			tags[trace.TAG_VERSION] = span.Meta[trace.VERSION]
			if tags[trace.TAG_VERSION] == "" {
				tags[trace.TAG_VERSION] = trace.GetFromPluginTag(DdtraceTags, trace.VERSION)
			}
			tags[trace.TAG_CONTAINER_HOST] = span.Meta[trace.CONTAINER_HOST]
			tags[trace.TAG_HTTP_METHOD] = span.Meta["http.method"]
			tags[trace.TAG_HTTP_CODE] = span.Meta["http.status_code"]

			for k, v := range DdtraceTags {
				tags[k] = v
			}

			// run trace data sample
			if !traceSampleConf.SampleFilter(tags[trace.TAG_SPAN_STATUS], tags, fmt.Sprintf("%d", span.TraceID)) {
				continue
			}

			field[trace.FIELD_RESOURCE] = span.Resource
			field[trace.FIELD_PARENTID] = fmt.Sprintf("%d", span.ParentID)
			field[trace.FIELD_TRACEID] = fmt.Sprintf("%d", span.TraceID)
			field[trace.FIELD_SPANID] = fmt.Sprintf("%d", span.SpanID)
			field[trace.FIELD_DURATION] = span.Duration / 1000
			field[trace.FIELD_START] = span.Start / 1000
			if v, ok := span.Metrics["system.pid"]; ok {
				field[trace.FIELD_PID] = fmt.Sprintf("%v", v)
			}

			js, err := json.Marshal(span)
			if err != nil {
				return err
			}
			field[trace.FIELD_MSG] = string(js)

			tm.Tags = tags
			tm.Fields = field
			tm.Ts = time.Unix(span.Start/int64(time.Second), span.Start%int64(time.Second))

			pt, err := tm.LineProto()
			if err != nil {
				return err
			}

			pts = append(pts, pt)
		}
	}

	return dkio.Feed(inputName, datakit.Tracing, pts, &dkio.Option{HighFreq: true})
}

func unmarshalDdtraceMsgpack(body io.ReadCloser) ([][]*Span, error) {
	var traces [][]*Span
	var mh codec.MsgpackHandle
	dec := codec.NewDecoder(body, &mh)
	err := dec.Decode(&traces)
	return traces, err
}

func getSpanAndParentId(spans []*Span) (map[uint64]string, map[uint64]string) {
	spanID := make(map[uint64]string)
	parentId := make(map[uint64]string)
	for _, span := range spans {
		if span == nil {
			continue
		}
		spanID[span.SpanID] = span.Service

		if span.ParentID != 0 {
			parentId[span.ParentID] = ""
		}
	}
	return spanID, parentId
}
