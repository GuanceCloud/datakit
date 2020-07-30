package trace

import (
	"net/http"
	"fmt"
	"encoding/json"
)

type SkyWalkTag struct {
	Key   string                `json:"key,omitempty"`
	Value interface{}           `json:"value,omitempty"`
}

type SkyWalkLogData struct {
	Key   string                `json:"key,omitempty"`
	Value interface{}           `json:"value,omitempty"`
}

type SkyWalkLog struct {
	Time int64
	Data []*SkyWalkLogData      `json:"data,omitempty"`
}

type SkyWalkRef struct {
	RefType                  int
	TraceId                  string
	ParentTraceSegmentId     string
	ParentSpanId             int
	ParentService            string
	ParentServiceInstance    string
	ParentEndpoint           string
	NetworkAddressUsedAtPeer string
}

type SkyWalkSpan struct {
	SpanId        uint64        `json:"spanId"`
	ParentSpanId  int64         `json:"parentSpanId"`
	StartTime     int64		    `json:"startTime"`
	EndTime       int64         `json:"endTime"`
	OperationName string        `json:"operationName"`
	Peer          string        `json:"peer"`
	SpanType      string        `json:"spanType"`
	SpanLayer     string        `json:"spanLayer"`
	ComponentId   uint64        `json:"componentId"`
	IsError       bool          `json:"isError"`
	Logs		  []*SkyWalkLog `json:"logs,omitempty"`
	Tags          []*SkyWalkTag `json:"tags,omitempty"`
	Refs		  []*SkyWalkRef `json:"Refs,omitempty"`
}

type SkyWalkSegment struct {
	TraceId         string
	TraceSegmentId  string
	Service         string
	ServiceInstance string
	Spans           []*SkyWalkSpan
}

const (
	SKYWALK_SEGMENT    = "/v3/segment"
	SKYWALK_SEGMENTS   = "/v3/segments"
	SKYWALK_PROPERTIES = "/v3/management/reportProperties"
	SKYWALK_KEEPALIVE  = "/v3/management/keepAlive"
)

func handleSkyWalking(w http.ResponseWriter, r *http.Request, path string, body []byte) error{
	log.Debugf("path = %s body = ->|%s|<-", path, string(body))
	switch path{
	case SKYWALK_SEGMENT:
		return handleSkyWalkSegment(w, r, body)
	case SKYWALK_SEGMENTS:
		return handleSkyWalkSegments(w, r, body)
	case SKYWALK_PROPERTIES:
		return handleSkyWalkProperties(w, r, body)
	case SKYWALK_KEEPALIVE:
		return handleSkyWalkKeepAlive(w, r, body)
	default:
		err := fmt.Errorf("skywalking path %s not founded", path)
		JsonReply(w, http.StatusNotFound, "%s", err)
		return err
	}
}

func skywalkToLineProto(sg *SkyWalkSegment) error {
	for _, span := range sg.Spans {
		t := TraceAdapter{}

		t.source = "skywalking"

		t.duration = (span.EndTime -span.StartTime)*1000
		t.timestampUs = span.StartTime * 1000
		js ,err := json.Marshal(span)
		if err != nil {
			return err
		}
		t.content = string(js)
		t.class         = "tracing"
		t.serviceName   = sg.Service
		t.operationName = span.OperationName
		if span.SpanType == "Entry" {
			if len(span.Refs) > 0 {
				t.parentID      = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			t.parentID      = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.ParentSpanId)
		}

		t.traceID       = sg.TraceId
		t.spanID        = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.SpanId)
		if span.IsError {
			t.isError   = "true"
		}
		if span.SpanType == "Entry" {
			t.spanType  = SPAN_TYPE_ENTRY
		} else {
			t.spanType  = SPAN_TYPE_LOCAL
		}
		t.endPoint      = span.Peer

		t.mkLineProto()
	}
	return nil
}
func handleSkyWalkSegment(w http.ResponseWriter, r *http.Request, body []byte) error {
	var sg SkyWalkSegment
	err := json.Unmarshal(body, &sg)
	if err != nil {
		return err
	}

	return skywalkToLineProto(&sg)
}

func handleSkyWalkSegments(w http.ResponseWriter, r *http.Request, body []byte) error {
	var sgs []SkyWalkSegment
	err := json.Unmarshal(body, &sgs)
	if err != nil {
		return err
	}

	for _, sg := range sgs {
		err := skywalkToLineProto(&sg)
		if err != nil {
			return nil
		}
	}
	return nil
}

func handleSkyWalkProperties(w http.ResponseWriter, r *http.Request, body []byte) error {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
	return nil
}

func handleSkyWalkKeepAlive(w http.ResponseWriter, r *http.Request, body []byte) error {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
	return nil
}