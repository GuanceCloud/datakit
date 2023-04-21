// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace convert tracing data from multiple platforms into datakit trace data structure
package trace

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

// tracing data constants
// nolint:stylecheck
const (
	// datakit tracing customer tags.
	UNKNOWN_SERVICE = "unknown_service"
	CONTAINER_HOST  = "container_host"
	ENV             = "env"
	PROJECT         = "project"
	VERSION         = "version"

	// span status.
	STATUS_OK       = "ok"
	STATUS_INFO     = "info"
	STATUS_DEBUG    = "debug"
	STATUS_WARN     = "warning"
	STATUS_ERR      = "error"
	STATUS_CRITICAL = "critical"

	// span position in trace.
	SPAN_TYPE_ENTRY   = "entry"
	SPAN_TYPE_LOCAL   = "local"
	SPAN_TYPE_EXIT    = "exit"
	SPAN_TYPE_UNKNOWN = "unknown"

	// span source type.
	SPAN_SOURCE_APP       = "app"
	SPAN_SOURCE_FRAMEWORK = "framework"
	SPAN_SOURCE_CACHE     = "cache"
	SPAN_SOURCE_MSGQUE    = "message_queue"
	SPAN_SOURCE_CUSTOMER  = "custom"
	SPAN_SOURCE_DB        = "db"
	SPAN_SOURCE_WEB       = "web"

	// line protocol tags.
	TAG_CONTAINER_HOST   = "container_host"
	TAG_ENDPOINT         = "endpoint"
	TAG_ENV              = "env"
	TAG_HTTP_HOST        = "http_host"
	TAG_HTTP_METHOD      = "http_method"
	TAG_HTTP_ROUTE       = "http_route"
	TAG_HTTP_STATUS_CODE = "http_status_code"
	TAG_HTTP_URL         = "http_url"
	TAG_OPERATION        = "operation"
	TAG_PID              = "pid"
	TAG_PROJECT          = "project"
	TAG_SERVICE          = "service"
	TAG_SOURCE_TYPE      = "source_type"
	TAG_SPAN_STATUS      = "status"
	TAG_SPAN_TYPE        = "span_type"
	TAG_VERSION          = "version"

	// line protocol fields.
	FIELD_DURATION    = "duration"
	FIELD_MESSAGE     = "message"
	FIELD_PARENTID    = "parent_id"
	FIELD_PRIORITY    = "priority"
	FIELD_RESOURCE    = "resource"
	FIELD_SAMPLE_RATE = "sample_rate"
	FIELD_SPANID      = "span_id"
	FIELD_START       = "start"
	FIELD_TRACEID     = "trace_id"
	FIELD_ERR_MESSAGE = "error_message"
	FIELD_ERR_STACK   = "error_stack"
	FIELD_ERR_TYPE    = "error_type"
	FIELD_CALL_TREE   = "calling_tree"
)

// nolint:stylecheck
const (
	// PriorityRuleSamplerReject specifies that the rule sampler has decided that this trace should be rejected.
	PRIORITY_RULE_SAMPLER_REJECT = -3
	// PriorityUserReject informs the backend that a trace should be rejected and not stored.
	// This should be used by user code overriding default priority.
	PRIORITY_USER_REJECT = -1
	// PriorityAutoReject informs the backend that a trace should be rejected and not stored.
	// This is used by the builtin sampler.
	PRIORITY_AUTO_REJECT = 0
	// PriorityAutoKeep informs the backend that a trace should be kept and not stored.
	// This is used by the builtin sampler.
	PRIORITY_AUTO_KEEP = 1
	// PriorityUserKeep informs the backend that a trace should be kept and not stored.
	// This should be used by user code overriding default priority.
	PRIORITY_USER_KEEP = 2
	// PriorityRuleSamplerKeep specifies that the rule sampler has decided that this trace should be kept.
	PRIORITY_RULE_SAMPLER_KEEP = 3
)

var (
	sourceTypes = map[string]string{
		"consul": SPAN_SOURCE_APP,

		".net":        SPAN_SOURCE_FRAMEWORK,
		"datanucleus": SPAN_SOURCE_FRAMEWORK,
		"django":      SPAN_SOURCE_FRAMEWORK,
		"express":     SPAN_SOURCE_FRAMEWORK,
		"flask":       SPAN_SOURCE_FRAMEWORK,
		"go-gin":      SPAN_SOURCE_FRAMEWORK,
		"graphql":     SPAN_SOURCE_FRAMEWORK,
		"hibernate":   SPAN_SOURCE_FRAMEWORK,
		"laravel":     SPAN_SOURCE_FRAMEWORK,
		"soap":        SPAN_SOURCE_FRAMEWORK,
		"spring":      SPAN_SOURCE_FRAMEWORK,

		"cache":     SPAN_SOURCE_CACHE,
		"memcached": SPAN_SOURCE_CACHE,
		"redis":     SPAN_SOURCE_CACHE,

		"aerospike":     SPAN_SOURCE_DB,
		"cassandra":     SPAN_SOURCE_DB,
		"db":            SPAN_SOURCE_DB,
		"elasticsearch": SPAN_SOURCE_DB,
		"influxdb":      SPAN_SOURCE_DB,
		"leveldb":       SPAN_SOURCE_DB,
		"mongodb":       SPAN_SOURCE_DB,
		"mysql":         SPAN_SOURCE_DB,
		"pymysql":       SPAN_SOURCE_DB,
		"sql":           SPAN_SOURCE_DB,

		"go-nsq":   SPAN_SOURCE_MSGQUE,
		"kafka":    SPAN_SOURCE_MSGQUE,
		"mqtt":     SPAN_SOURCE_MSGQUE,
		"queue":    SPAN_SOURCE_MSGQUE,
		"rabbitmq": SPAN_SOURCE_MSGQUE,
		"rocketmq": SPAN_SOURCE_MSGQUE,

		"dns":   SPAN_SOURCE_WEB,
		"grpc":  SPAN_SOURCE_WEB,
		"http":  SPAN_SOURCE_WEB,
		"http2": SPAN_SOURCE_WEB,
		"rpc":   SPAN_SOURCE_WEB,
		"web":   SPAN_SOURCE_WEB,

		"":          SPAN_SOURCE_CUSTOMER,
		"benchmark": SPAN_SOURCE_CUSTOMER,
		"build":     SPAN_SOURCE_CUSTOMER,
		"custom":    SPAN_SOURCE_CUSTOMER,
		"template":  SPAN_SOURCE_CUSTOMER,
		"test":      SPAN_SOURCE_CUSTOMER,
		"worker":    SPAN_SOURCE_CUSTOMER,
	}
	priorityRules = map[int]string{
		PRIORITY_RULE_SAMPLER_REJECT: "PRIORITY_RULE_SAMPLER_REJECT",
		PRIORITY_USER_REJECT:         "PRIORITY_USER_REJECT",
		PRIORITY_AUTO_REJECT:         "PRIORITY_AUTO_REJECT",
		PRIORITY_AUTO_KEEP:           "PRIORITY_AUTO_KEEP",
		PRIORITY_USER_KEEP:           "PRIORITY_USER_KEEP",
		PRIORITY_RULE_SAMPLER_KEEP:   "PRIORITY_RULE_SAMPLER_KEEP",
	}
)

func GetSpanSourceType(app string) string {
	if s, ok := sourceTypes[strings.ToLower(app)]; ok {
		return s
	} else {
		return SPAN_SOURCE_CUSTOMER
	}
}

type DatakitSpan struct {
	TraceID    string                 `json:"trace_id"`
	ParentID   string                 `json:"parent_id"`
	SpanID     string                 `json:"span_id"`
	Service    string                 `json:"service"`     // service name
	Resource   string                 `json:"resource"`    // resource or api under service
	Operation  string                 `json:"operation"`   // api name
	Source     string                 `json:"source"`      // client tracer name
	SpanType   string                 `json:"span_type"`   // relative span position in tracing: entry, local, exit or unknow
	SourceType string                 `json:"source_type"` // service type
	Tags       map[string]string      `json:"tags"`
	Metrics    map[string]interface{} `json:"metrics"`
	Start      int64                  `json:"start"`    // unit: nano sec
	Duration   int64                  `json:"duration"` // unit: nano sec
	Status     string                 `json:"status"`   // span status like error, ok, info etc.
	Content    string                 `json:"content"`  // raw tracing data in json
}

type DatakitTrace []*DatakitSpan

type DatakitTraces []DatakitTrace

func FindSpanTypeInMultiServersIntSpanID(spanID, parentID uint64, service string, spanIDs map[uint64]string, parentIDs map[uint64]bool) string {
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

func FindSpanTypeIntSpanID(spanID, parentID uint64, spanIDs, parentIDs map[uint64]bool) string {
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

func UnifyToUint64ID(id string) uint64 {
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
		i   uint64
		err error
	)
	if isInt {
		if i, err = strconv.ParseUint(id, 10, 64); err == nil {
			return i
		}
	}
	if isHex {
		if i, err = strconv.ParseUint(id, 16, 64); err == nil {
			return i
		}
	}

	hexstr := hex.EncodeToString([]byte(id))
	if l := len(hexstr); l > 16 {
		hexstr = hexstr[l-16:]
	}
	i, _ = strconv.ParseUint(hexstr, 16, 64)

	return i
}

func MergeInToCustomerTags(customerKeys []string, datakitTags, sourceTags map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range datakitTags {
		merged[k] = v
	}
	for _, k := range customerKeys {
		if v, ok := sourceTags[k]; ok {
			merged[k] = v
		}
	}

	return merged
}

// ParseTracerRequest parse the given http request to Content-Type and body buffer if no error
// occurred. If the given body in request is compressed by gzip, decompression work will
// be done automatically.
func ParseTracerRequest(req *http.Request) (contentType, encode string, buf []byte, err error) {
	if req == nil {
		err = errors.New("nil http.Request pointer")

		return
	}

	var body io.ReadCloser
	if ihttp.GetHeader(req, "Content-Encoding") == "gzip" {
		encode = "gzip"
		if body, err = gzip.NewReader(req.Body); err == nil {
			defer body.Close() // nolint:errcheck
		}
	} else {
		body = req.Body
	}

	if buf, err = io.ReadAll(body); err != nil {
		return
	}

	contentType = ihttp.GetHeader(req, "Content-Type")

	return
}

type TraceParameters struct {
	URLPath string
	Media   string
	Encode  string
	Body    *bytes.Buffer
}

func MergeTags(input ...map[string]string) map[string]string {
	tags := make(map[string]string)
	for i := range input {
		for k, v := range input[i] {
			tags[k] = v
		}
	}

	return tags
}

func MergeFields(input ...map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := range input {
		for k, v := range input[i] {
			fields[k] = v
		}
	}

	return fields
}
