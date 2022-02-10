// Package trace convert tracing data from multiple platforms into datakit trace data structure
package trace

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

//nolint:stylecheck
const (
	CONTAINER_HOST = "container_host"
	ENV            = "env"
	PROJECT        = "project"
	VERSION        = "version"

	STATUS_OK       = "ok"
	STATUS_INFO     = "info"
	STATUS_WARN     = "warning"
	STATUS_ERR      = "error"
	STATUS_CRITICAL = "critical"

	SPAN_TYPE_ENTRY = "entry"
	SPAN_TYPE_LOCAL = "local"
	SPAN_TYPE_EXIT  = "exit"

	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"

	TAG_CONTAINER_HOST = "container_host"
	TAG_ENDPOINT       = "endpoint"
	TAG_ENV            = "env"
	TAG_HTTP_CODE      = "http_status_code"
	TAG_HTTP_METHOD    = "http_method"
	TAG_OPERATION      = "operation"
	TAG_PROJECT        = "project"
	TAG_SERVICE        = "service"
	TAG_SPAN_STATUS    = "status"
	TAG_SPAN_TYPE      = "span_type"
	TAG_TYPE           = "type"
	TAG_VERSION        = "version"

	FIELD_DURATION = "duration"
	FIELD_MSG      = "message"
	FIELD_PARENTID = "parent_id"
	FIELD_PID      = "pid"
	FIELD_RESOURCE = "resource"
	FIELD_SPANID   = "span_id"
	FIELD_START    = "start"
	FIELD_TRACEID  = "trace_id"
)

// trace keep priority
const (
	// reject trace before send to dataway
	PriorityReject = -1
	// consider with sampling rate
	PriorityAuto = 0
	// always send to dataway and do not consider sampling and filters
	PriorityKeep = 1
)

var (
	packageName = "dktrace"
	log         = logger.DefaultSLogger(packageName)
)

type DatakitSpan struct {
	TraceID            string
	ParentID           string
	SpanID             string
	Service            string
	Resource           string
	Operation          string
	Source             string // third part source name
	SpanType           string
	SourceType         string
	Env                string
	Project            string
	Version            string
	Tags               map[string]string
	EndPoint           string
	HTTPMethod         string
	HTTPStatusCode     string
	ContainerHost      string
	PID                string // process id
	Start              int64  // nano sec
	Duration           int64  // nano sec
	Status             string
	Content            string
	Priority           int
	SamplingRateGlobal float64
}

type DatakitTrace []*DatakitSpan

type DatakitTraces []DatakitTrace

func FindSpanTypeInt(spanID, parentID int64, spanIDs, parentIDs map[int64]bool) string {
	if parentID != 0 {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func FindSpanTypeString(spanID, parentID string, spanIDs, parentIDs map[string]bool) string {
	if parentID != "" && parentID != "0" {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func MergeTags(data ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, tags := range data {
		for k, v := range tags {
			merged[k] = v
		}
	}

	return merged
}

func GetTraceInt64ID(high, low int64) int64 {
	temp := low
	for temp != 0 {
		high *= 10
		temp /= 10
	}

	return high + low
}

func GetTraceStringID(high, low int64) string {
	return fmt.Sprintf("%d%d", high, low)
}

func IsRootSpan(dkspan *DatakitSpan) bool {
	return dkspan.ParentID == "0" || dkspan.ParentID == ""
}

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
	Body        []byte
}

func ParseTraceInfo(req *http.Request) (*TraceReqInfo, error) {
	defer req.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	reqInfo := &TraceReqInfo{
		Source:      req.URL.Query().Get("source"),
		ContentType: req.Header.Get("Content-Type"),
		Version:     req.URL.Query().Get("version"),
		Body:        body,
	}
	if req.Header.Get("Content-Encoding") == "gzip" {
		var rd *gzip.Reader
		if rd, err = gzip.NewReader(bytes.NewBuffer(body)); err == nil {
			if body, err = io.ReadAll(rd); err == nil {
				reqInfo.Body = body
			}
		}
	}

	return reqInfo, err
}

func UnifyTraceIDInt64(traceID string) int64 {
	if len(traceID) == 0 {
		return 0
	}

	var isAllInt = true
	for _, b := range traceID {
		if b < 48 || b > 57 {
			isAllInt = false
			break
		}
	}

	if isAllInt {
		if i, err := strconv.ParseInt(traceID, 10, 64); err == nil {
			return i
		}
	}

	hexstr := hex.EncodeToString([]byte(traceID))
	if l := len(hexstr); l > 16 {
		hexstr = hexstr[l-16:]
	}
	i, _ := strconv.ParseInt(hexstr, 16, 64)

	return i
}
