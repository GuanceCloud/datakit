package traceSkywalking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type SkyWalkTag struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type SkyWalkLogData struct {
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type SkyWalkLog struct {
	Time float64
	Data []*SkyWalkLogData `json:"data,omitempty"`
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
	StartTime     int64         `json:"startTime"`
	EndTime       int64         `json:"endTime"`
	OperationName string        `json:"operationName"`
	Peer          string        `json:"peer"`
	SpanType      string        `json:"spanType"`
	SpanLayer     string        `json:"spanLayer"`
	ComponentId   uint64        `json:"componentId"`
	IsError       bool          `json:"isError"`
	Logs          []*SkyWalkLog `json:"logs,omitempty"`
	Tags          []*SkyWalkTag `json:"tags,omitempty"`
	Refs          []*SkyWalkRef `json:"Refs,omitempty"`
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

func SkywalkingTraceHandle(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleSkywalkingTrace(w, r); err != nil {
		io.FeedLastError(inputName, err.Error())
		log.Errorf("%v", err)
	}
}

func handleSkywalkingTrace(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	return handleSkyWalking(w, r, r.URL.Path, reqInfo.Body)
}

func handleSkyWalking(w http.ResponseWriter, r *http.Request, path string, body []byte) error {
	log.Debugf("path = %s body = ->|%s|<-", path, string(body))
	switch path {
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
		return err
	}
}

func skywalkToLineProto(sg *SkyWalkSegment) error {
	adapterGroup := []*trace.TraceAdapter{}
	for _, span := range sg.Spans {
		tAdapter := &trace.TraceAdapter{}

		tAdapter.Source = "skywalking"

		tAdapter.Duration = (span.EndTime - span.StartTime) * 1000000
		tAdapter.Start = span.StartTime * 1000000
		js, err := json.Marshal(span)
		if err != nil {
			return err
		}
		tAdapter.Content = string(js)
		tAdapter.ServiceName = sg.Service
		tAdapter.OperationName = span.OperationName
		if span.SpanType == "Entry" {
			if len(span.Refs) > 0 {
				tAdapter.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			tAdapter.ParentID = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.ParentSpanId)
		}

		tAdapter.TraceID = sg.TraceId
		tAdapter.SpanID = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.SpanId)
		tAdapter.Status = trace.STATUS_OK
		if span.IsError {
			tAdapter.Status = trace.STATUS_ERR
		}
		if span.SpanType == "Entry" {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		} else if span.SpanType == "Exit" {
			tAdapter.SpanType = trace.SPAN_TYPE_EXIT
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		}
		tAdapter.EndPoint = span.Peer
		tAdapter.Tags = SkywalkingTagsV3

		// run trace data sample
		if traceSampleConf.SampleFilter(tAdapter.Status, tAdapter.Tags, tAdapter.TraceID) {
			adapterGroup = append(adapterGroup, tAdapter)
		}
	}
	trace.MkLineProto(adapterGroup, inputName)

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
