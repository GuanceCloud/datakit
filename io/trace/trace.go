// Package trace convert tracing data from multiple platforms into datakit trace data structure
package trace

import (
	"compress/gzip"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

// tracing data constants
// nolint:stylecheck
const (
	// datakit tracing customer tags.
	CONTAINER_HOST = "container_host"
	ENV            = "env"
	PROJECT        = "project"
	VERSION        = "version"

	// span status.
	STATUS_OK       = "ok"
	STATUS_INFO     = "info"
	STATUS_WARN     = "warning"
	STATUS_ERR      = "error"
	STATUS_CRITICAL = "critical"

	// span position in trace.
	SPAN_TYPE_ENTRY  = "entry"
	SPAN_TYPE_LOCAL  = "local"
	SPAN_TYPE_EXIT   = "exit"
	SPAN_TYPE_UNKNOW = "unknow"

	// service type.
	SPAN_SERVICE_APP    = "app"
	SPAN_SERVICE_CACHE  = "cache"
	SPAN_SERVICE_CUSTOM = "custom"
	SPAN_SERVICE_DB     = "db"
	SPAN_SERVICE_WEB    = "web"

	// line protocol tags.
	TAG_CONTAINER_HOST = "container_host"
	TAG_ENDPOINT       = "endpoint"
	TAG_ENV            = "env"
	TAG_HTTP_CODE      = "http_status_code"
	TAG_HTTP_METHOD    = "http_method"
	TAG_OPERATION      = "operation"
	TAG_PROJECT        = "project"
	TAG_SERVICE        = "service"
	TAG_SOURCE_TYPE    = "source_type"
	TAG_SPAN_STATUS    = "status"
	TAG_SPAN_TYPE      = "span_type"
	TAG_VERSION        = "version"

	// line protocol fields.
	FIELD_DURATION           = "duration"
	FIELD_MSG                = "message"
	FIELD_PARENTID           = "parent_id"
	FIELD_PID                = "pid"
	FIELD_PRIORITY           = "priority"
	FIELD_RESOURCE           = "resource"
	FIELD_SAMPLE_RATE_GLOBAL = "sample_rate_global"
	FIELD_SPANID             = "span_id"
	FIELD_START              = "start"
	FIELD_TRACEID            = "trace_id"
)

var (
	packageName = "dktrace"
	log         = logger.DefaultSLogger(packageName)
)

type DatakitSpan struct {
	TraceID            string            `json:"trace_id"`
	ParentID           string            `json:"parent_id"`
	SpanID             string            `json:"span_id"`
	Service            string            `json:"service"`     // process name
	Resource           string            `json:"resource"`    // a resource name in process
	Operation          string            `json:"operation"`   // a operation name behind resource
	Source             string            `json:"source"`      // tracer name
	SpanType           string            `json:"span_type"`   // span type of entry, local, exit or unknow
	SourceType         string            `json:"source_type"` // process role in service
	Env                string            `json:"env"`
	Project            string            `json:"project"`
	Version            string            `json:"version"`
	Tags               map[string]string `json:"tags"`
	EndPoint           string            `json:"end_point"`
	HTTPMethod         string            `json:"http_method"`
	HTTPStatusCode     string            `json:"http_status_code"`
	ContainerHost      string            `json:"container_host"`
	PID                string            `json:"p_id"`     // process id
	Start              int64             `json:"start"`    // unit: nano sec
	Duration           int64             `json:"duration"` // unit: nano sec
	Status             string            `json:"status"`
	Content            string            `json:"content"`              // raw tracing data in json
	Priority           int               `json:"priority"`             // smapling priority
	SamplingRateGlobal float64           `json:"sampling_rate_global"` // global sampling ratio
}

type DatakitTrace []*DatakitSpan

type DatakitTraces []DatakitTrace

func FindSpanTypeInMultiServersIntSpanID(spanID, parentID int64, service string, spanIDs map[int64]string, parentIDs map[int64]bool) string {
	if parentID != 0 {
		if ss, ok := spanIDs[parentID]; ok {
			if service != ss {
				return SPAN_TYPE_ENTRY
			}
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func FindSpanTypeInMultiServersStrSpanID(spanID, parentID string, service string, spanIDs map[string]string, parentIDs map[string]bool) string {
	if parentID != "0" && parentID != "" {
		if ss, ok := spanIDs[parentID]; ok {
			if service != ss {
				return SPAN_TYPE_ENTRY
			}
			if parentIDs[spanID] {
				return SPAN_TYPE_LOCAL
			} else {
				return SPAN_TYPE_EXIT
			}
		}
	}

	return SPAN_TYPE_ENTRY
}

func FindSpanTypeIntSpanID(spanID, parentID int64, spanIDs, parentIDs map[int64]bool) string {
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

func FindSpanTypeStrSpanID(spanID, parentID string, spanIDs, parentIDs map[string]bool) string {
	if parentID != "0" && parentID != "" {
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

func GetTraceInt64ID(high, low int64) int64 {
	temp := low
	for temp != 0 {
		high *= 10
		temp /= 10
	}

	return high + low
}

func IsRootSpan(dkspan *DatakitSpan) bool {
	return dkspan.ParentID == "0" || dkspan.ParentID == ""
}

func UnifyToInt64ID(id string) int64 {
	if len(id) == 0 {
		return 0
	}

	var (
		isInt = true
		isHex = true
	)
	for _, b := range id {
		if b < '0' || b > '9' {
			isInt = false
			if b < 'a' || b > 'f' {
				isHex = false
				break
			}
		}
	}
	var (
		i   int64
		err error
	)
	if isInt {
		if i, err = strconv.ParseInt(id, 10, 64); err == nil {
			return i
		}
	}
	if isHex {
		if i, err = strconv.ParseInt(id, 16, 64); err == nil {
			return i
		}
	}

	hexstr := hex.EncodeToString([]byte(id))
	if l := len(hexstr); l > 16 {
		hexstr = hexstr[l-16:]
	}
	i, _ = strconv.ParseInt(hexstr, 16, 64)

	return i
}

func MergeInToCustomerTags(customerKeys []string, datakitTags, sourceTags map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range datakitTags {
		merged[k] = v
	}
	for i := range customerKeys {
		if v, ok := sourceTags[customerKeys[i]]; ok {
			merged[customerKeys[i]] = v
		}
	}

	return merged
}

func ParseTracingRequest(req *http.Request) (contentType string, body io.ReadCloser, err error) {
	if req == nil {
		return "", nil, errors.New("nil http.Request pointer")
	}

	contentType = ihttp.GetHeader(req, "Content-Type")
	if ihttp.GetHeader(req, "Content-Encoding") == "gzip" {
		body, err = gzip.NewReader(req.Body)
	} else {
		body = req.Body
	}

	return
}
