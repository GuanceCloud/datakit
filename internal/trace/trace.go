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
	"regexp"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"google.golang.org/grpc"
)

// tracing data constants
// nolint:stylecheck
const (
	// datakit tracing customer tags.
	UnknownService = "unknown_service"
	ContainerHost  = "container_host"
	Env            = "env"
	Project        = "project"
	Version        = "version"

	StatusOk       = "ok"
	StatusInfo     = "info"
	StatusDebug    = "debug"
	StatusWarn     = "warning"
	StatusErr      = "error"
	StatusCritical = "critical"

	// span position in trace.
	SpanTypeEntry   = "entry"
	SpanTypeLocal   = "local"
	SpanTypeExit    = "exit"
	SpanTypeUnknown = "unknown"

	// span source type.
	SpanSourceApp       = "app"
	SpanSourceFramework = "framework"
	SpanSourceCache     = "cache"
	SpanSourceMsgque    = "message_queue"
	SpanSourceCustomer  = "custom"
	SpanSourceDb        = "db"
	SpanSourceWeb       = "web"

	// line protocol tags.
	TagHost           = "host"
	TagContainerHost  = "container_host"
	TagEndpoint       = "endpoint"
	TagEnv            = "env"
	TagHttpHost       = "http_host"
	TagHttpMethod     = "http_method"
	TagHttpRoute      = "http_route"
	TagHttpStatusCode = "http_status_code"
	TagHttpUrl        = "http_url"
	TagOperation      = "operation"
	TagSource         = "source"
	TagPid            = "pid"
	TagProject        = "project"
	TagService        = "service"
	TagSourceType     = "source_type"
	TagSpanStatus     = "status"
	TagSpanType       = "span_type"
	TagVersion        = "version"

	// line protocol fields.
	FieldDuration   = "duration"
	FieldMessage    = "message"
	FieldParentID   = "parent_id"
	FieldResource   = "resource"
	FieldSpanid     = "span_id"
	FieldStart      = "start"
	FieldTraceID    = "trace_id"
	FieldErrMessage = "error_message"
	FieldErrStack   = "error_stack"
	FieldErrType    = "error_type"
	FieldCallTree   = "calling_tree"
	Trace128BitId   = "trace_128_bit_id"
)

// nolint:stylecheck
const (
	// PriorityRuleSamplerReject specifies that the rule sampler has decided that this trace should be rejected.
	PriorityRuleSamplerReject = -3
	// PriorityUserReject informs the backend that a trace should be rejected and not stored.
	// This should be used by user code overriding default priority.
	PriorityUserReject = -1
	// PriorityAutoReject informs the backend that a trace should be rejected and not stored.
	// This is used by the builtin sampler.
	PriorityAutoReject = 0
	// PriorityAutoKeep informs the backend that a trace should be kept and not stored.
	// This is used by the builtin sampler.
	PriorityAutoKeep = 1
	// PriorityUserKeep informs the backend that a trace should be kept and not stored.
	// This should be used by user code overriding default priority.
	PriorityUserKeep = 2
	// PriorityRuleSamplerKeep specifies that the rule sampler has decided that this trace should be kept.
	PriorityRuleSamplerKeep = 3
)

var (
	sourceTypes = map[string]string{
		"consul": SpanSourceApp,

		".net":        SpanSourceFramework,
		"datanucleus": SpanSourceFramework,
		"django":      SpanSourceFramework,
		"express":     SpanSourceFramework,
		"flask":       SpanSourceFramework,
		"go-gin":      SpanSourceFramework,
		"graphql":     SpanSourceFramework,
		"hibernate":   SpanSourceFramework,
		"laravel":     SpanSourceFramework,
		"soap":        SpanSourceFramework,
		"spring":      SpanSourceFramework,

		"cache":     SpanSourceCache,
		"memcached": SpanSourceCache,
		"redis":     SpanSourceCache,

		"aerospike":     SpanSourceDb,
		"cassandra":     SpanSourceDb,
		"db":            SpanSourceDb,
		"elasticsearch": SpanSourceDb,
		"influxdb":      SpanSourceDb,
		"leveldb":       SpanSourceDb,
		"mongodb":       SpanSourceDb,
		"mysql":         SpanSourceDb,
		"pymysql":       SpanSourceDb,
		"sql":           SpanSourceDb,

		"go-nsq":   SpanSourceMsgque,
		"kafka":    SpanSourceMsgque,
		"mqtt":     SpanSourceMsgque,
		"queue":    SpanSourceMsgque,
		"rabbitmq": SpanSourceMsgque,
		"rocketmq": SpanSourceMsgque,

		"dns":   SpanSourceWeb,
		"grpc":  SpanSourceWeb,
		"http":  SpanSourceWeb,
		"http2": SpanSourceWeb,
		"rpc":   SpanSourceWeb,
		"web":   SpanSourceWeb,

		"":          SpanSourceCustomer,
		"benchmark": SpanSourceCustomer,
		"build":     SpanSourceCustomer,
		"custom":    SpanSourceCustomer,
		"template":  SpanSourceCustomer,
		"test":      SpanSourceCustomer,
		"worker":    SpanSourceCustomer,
	}

	DefaultGRPCServerOpts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 10),
	}
)

func GetSpanSourceType(app string) string {
	if s, ok := sourceTypes[strings.ToLower(app)]; ok {
		return s
	} else {
		return SpanSourceCustomer
	}
}

func FindSpanTypeInMultiServersIntSpanID(spanID, parentID uint64, service string, spanIDs map[uint64]string, parentIDs map[uint64]bool) string {
	if parentID != 0 {
		if ss, ok := spanIDs[parentID]; ok {
			if service != ss {
				return SpanTypeEntry
			}
			if parentIDs[spanID] {
				return SpanTypeLocal
			} else {
				return SpanTypeExit
			}
		}
	}

	return SpanTypeEntry
}

func FindSpanTypeInMultiServersStrSpanID(spanID, parentID string, service string, spanIDs map[string]string, parentIDs map[string]bool) string {
	if parentID != "0" && parentID != "" {
		if ss, ok := spanIDs[parentID]; ok {
			if service != ss {
				return SpanTypeEntry
			}
			if parentIDs[spanID] {
				return SpanTypeLocal
			} else {
				return SpanTypeExit
			}
		}
	}

	return SpanTypeEntry
}

func FindSpanTypeIntSpanID(spanID, parentID uint64, spanIDs, parentIDs map[uint64]bool) string {
	if parentID != 0 {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SpanTypeLocal
			} else {
				return SpanTypeExit
			}
		}
	}

	return SpanTypeEntry
}

func FindSpanTypeStrSpanID(spanID, parentID string, spanIDs, parentIDs map[string]bool) string {
	if parentID != "0" && parentID != "" {
		if spanIDs[parentID] {
			if parentIDs[spanID] {
				return SpanTypeLocal
			} else {
				return SpanTypeExit
			}
		}
	}

	return SpanTypeEntry
}

func GetTraceInt64ID(high, low int64) int64 {
	temp := low
	for temp != 0 {
		high *= 10
		temp /= 10
	}

	return high + low
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

func MergeInToCustomerTags(dkTags, srcTags map[string]string, ignoreTags []*regexp.Regexp) (map[string]string, error) {
	merged := make(map[string]string)
	for k, v := range dkTags {
		merged[k] = v
	}
	for k, v := range srcTags {
		merged[replacer.Replace(k)] = v
	}

	for _, reg := range ignoreTags {
		for k := range merged {
			if reg.MatchString(k) {
				delete(merged, k)
			}
		}
	}

	return merged, nil
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
	if httpapi.GetHeader(req, "Content-Encoding") == "gzip" {
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

	contentType = httpapi.GetHeader(req, "Content-Type")

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
