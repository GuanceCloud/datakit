package ddtrace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/ugorji/go/codec"

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

	adapterGroup := []*trace.TraceAdapter{}
	for _, spans := range dspans {
		spanIds, parentIds := getSpanAndParentId(spans)
		for _, span := range spans {
			tAdpter := &trace.TraceAdapter{}
			tAdpter.Source = "ddtrace"

			if span.ParentID == 0 {
				tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
			} else {
				if serviceName, ok := spanIds[span.ParentID]; ok {
					if serviceName != span.Service {
						tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
					} else {
						if _, ok := parentIds[span.SpanID]; ok {
							tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
						} else {
							tAdpter.SpanType = trace.SPAN_TYPE_EXIT
						}
					}
				} else {
					tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
				}
			}

			tAdpter.ServiceName = span.Service
			tAdpter.OperationName = span.Name
			tAdpter.Resource = span.Resource

			tAdpter.ParentID = fmt.Sprintf("%d", span.ParentID)
			tAdpter.TraceID = fmt.Sprintf("%d", span.TraceID)
			tAdpter.SpanID = fmt.Sprintf("%d", span.SpanID)

			tAdpter.Type = ddtraceSpanType[span.Type]
			if span.Error == 0 {
				tAdpter.Status = trace.STATUS_OK
			} else {
				tAdpter.Status = trace.STATUS_ERR
			}

			tAdpter.Duration = span.Duration
			tAdpter.Start = span.Start

			if v, ok := span.Metrics["system.pid"]; ok {
				tAdpter.Pid = fmt.Sprintf("%v", v)
			}

			tAdpter.Project = span.Meta[trace.PROJECT]
			if tAdpter.Project == "" {
				tAdpter.Project = trace.GetFromPluginTag(DdtraceTags, trace.PROJECT)
			}

			tAdpter.Env = span.Meta[trace.ENV]
			if tAdpter.Env == "" {
				tAdpter.Env = trace.GetFromPluginTag(DdtraceTags, trace.ENV)
			}

			tAdpter.Version = span.Meta[trace.VERSION]
			if tAdpter.Version == "" {
				tAdpter.Version = trace.GetFromPluginTag(DdtraceTags, trace.VERSION)
			}

			js, err := json.Marshal(span)
			if err != nil {
				return err
			}
			tAdpter.Content = string(js)
			tAdpter.Tags = DdtraceTags

			adapterGroup = append(adapterGroup, tAdpter)
		}
	}

	trace.MkLineProto(adapterGroup, inputName)
	return nil
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
