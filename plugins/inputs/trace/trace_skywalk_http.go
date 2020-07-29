package trace

import (
	"net/http"
	"fmt"
	"encoding/json"
)

type SkyWalkTag struct {
	Key   string
	Value interface{}
}

type SkyWalkSpan struct {
	OperationName string       `json:"operationName,omitempty"`
	StartTime     int64		   `json:"startTime,omitempty"`
	EndTime       int64        `json:"endTime,omitempty"`
	Tags          []*SkyWalkTag `json:"tags,omitempty"`
	SpanType      string        `spanType:"tags,omitempty"`
	SpanId        uint64        `spanId:"tags,omitempty"`
	IsError       bool          `isError:"tags,omitempty"`
	ParentSpanId  int64         `parentSpanId:"tags,omitempty"`
	ComponentId   uint64        `componentId:"tags,omitempty"`
	Peer          string        `peer:"tags,omitempty"`
	SpanLayer     string        `spanLayer:"tags,omitempty"`
}

type SkyWalkSegment struct {
	TraceId         string
	ServiceInstance string
	Spans           []*SkyWalkSpan
	Service         string
	TraceSegmentId  string
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
		t.parentID      = fmt.Sprintf("%x", span.ParentSpanId)
		t.traceID       = sg.TraceId
		t.spanID        = fmt.Sprintf("%x", span.SpanId)
		if span.IsError {
			t.isError   = "true"
		}
		t.spanType      = span.SpanType
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