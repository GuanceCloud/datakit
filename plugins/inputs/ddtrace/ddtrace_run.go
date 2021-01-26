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
		for _, span := range spans {
			tAdpter := &trace.TraceAdapter{}
			tAdpter.Source = "ddtrace"

			tAdpter.ServiceName = span.Service
			tAdpter.OperationName = span.Name
			tAdpter.Resource = span.Resource

			tAdpter.ParentID = fmt.Sprintf("%d", span.ParentID)
			tAdpter.TraceID = fmt.Sprintf("%d", span.TraceID)
			tAdpter.SpanID = fmt.Sprintf("%d", span.SpanID)

			tAdpter.Type = span.Type
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
				tAdpter.Project = trace.GetProjectFromPluginTag(DdtraceTags)
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
