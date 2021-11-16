package skywalking

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type SkyWalkingTag struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type SkyWalkingLogData struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type SkyWalkingLog struct {
	Time float64
	Data []*SkyWalkingLogData `json:"data,omitempty"`
}

type SkyWalkingRef struct {
	RefType                  int
	TraceID                  string
	ParentTraceSegmentID     string
	ParentSpanID             int
	ParentService            string
	ParentServiceInstance    string
	ParentEndpoint           string
	NetworkAddressUsedAtPeer string
}

type SkyWalkingSpan struct {
	SpanID        uint64           `json:"spanId"`
	ParentSpanID  int64            `json:"parentSpanId"`
	StartTime     int64            `json:"startTime"`
	EndTime       int64            `json:"endTime"`
	OperationName string           `json:"operationName"`
	Peer          string           `json:"peer"`
	SpanType      string           `json:"spanType"`
	SpanLayer     string           `json:"spanLayer"`
	ComponentID   uint64           `json:"componentId"`
	IsError       bool             `json:"isError"`
	Logs          []*SkyWalkingLog `json:"logs,omitempty"`
	Tags          []*SkyWalkingTag `json:"tags,omitempty"`
	Refs          []*SkyWalkingRef `json:"refs,omitempty"`
}

type SkyWalkingSegment struct {
	TraceID         string
	TraceSegmentID  string
	Service         string
	ServiceInstance string
	Spans           []*SkyWalkingSpan
}

func handleSkyWalkingSegment(resp http.ResponseWriter, req *http.Request) {
	reqInfo, err := trace.ParseHTTPReq(req)
	if err != nil {
		log.Debug(err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	var segment SkyWalkingSegment
	if err = json.Unmarshal(reqInfo.Body, &segment); err != nil {
		log.Debug(err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	group, err := segmentToAdapters(&segment)
	if err != nil {
		log.Debug(err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}
	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Debug("empty segement")
	}

	resp.WriteHeader(http.StatusOK)
}

func handleSkyWalkingSegments(resp http.ResponseWriter, req *http.Request) {
	reqInfo, err := trace.ParseHTTPReq(req)
	if err != nil {
		log.Debug(err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	var segments []SkyWalkingSegment
	if err = json.Unmarshal(reqInfo.Body, &segments); err != nil {
		log.Debug(err)
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	var group []*trace.TraceAdapter
	for i := range segments {
		segment := &segments[i]
		group, err = segmentToAdapters(segment)
		if err != nil {
			log.Debug(err)
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
		if len(group) != 0 {
			trace.MkLineProto(group, inputName)
		} else {
			log.Debug("empty")
		}
	}

	resp.WriteHeader(http.StatusOK)
}

func handleSkyWalkingProperties(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleSkyWalkingKeepAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func segmentToAdapters(segment *SkyWalkingSegment) ([]*trace.TraceAdapter, error) {
	var adapterGroup []*trace.TraceAdapter
	for _, span := range segment.Spans {
		tAdapter := &trace.TraceAdapter{Source: inputName}
		tAdapter.Duration = (span.EndTime - span.StartTime) * 1000000
		tAdapter.Start = span.StartTime * 1000000
		js, err := json.Marshal(span)
		if err != nil {
			return adapterGroup, err
		}

		tAdapter.Content = string(js)
		tAdapter.ServiceName = segment.Service
		tAdapter.OperationName = span.OperationName
		if span.SpanType == "Entry" {
			if len(span.Refs) > 0 {
				tAdapter.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentID,
					span.Refs[0].ParentSpanID)
			}
		} else {
			tAdapter.ParentID = fmt.Sprintf("%s%d", segment.TraceSegmentID, span.ParentSpanID)
		}

		tAdapter.TraceID = segment.TraceID
		tAdapter.SpanID = fmt.Sprintf("%s%d", segment.TraceSegmentID, span.SpanID)
		tAdapter.Status = trace.STATUS_OK
		if span.IsError {
			tAdapter.Status = trace.STATUS_ERR
		}

		switch span.SpanType {
		case "Entry":
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		case "Exit":
			tAdapter.SpanType = trace.SPAN_TYPE_EXIT
		default:
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		}

		tAdapter.EndPoint = span.Peer
		tAdapter.Tags = skywalkingV3Tags

		adapterGroup = append(adapterGroup, tAdapter)
	}

	return adapterGroup, nil
}
