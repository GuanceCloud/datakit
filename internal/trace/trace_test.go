// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

var (
	_services     = []string{"login", "game", "fire_gun", "march", "kill", "logout"}
	_resources    = []string{"/get_user/name", "/push/data", "/check/security", "/fetch/data_source", "/pull/all_data", "/list/user_name"}
	_source       = []string{"ddtrace", "jaeger", "opentelemetry", "skywalking", "zipkin"}
	_span_types   = []string{SPAN_TYPE_ENTRY, SPAN_TYPE_LOCAL, SPAN_TYPE_EXIT, SPAN_TYPE_UNKNOW}
	_source_types = []string{SPAN_SERVICE_APP, SPAN_SERVICE_CACHE, SPAN_SERVICE_CUSTOM, SPAN_SERVICE_DB, SPAN_SERVICE_WEB}
	_http_methods = []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
	}
	_http_status_codes = []string{
		http.StatusText(http.StatusContinue),
		http.StatusText(http.StatusSwitchingProtocols),
		http.StatusText(http.StatusProcessing),
		http.StatusText(http.StatusEarlyHints),
		http.StatusText(http.StatusOK),
		http.StatusText(http.StatusCreated),
		http.StatusText(http.StatusAccepted),
		http.StatusText(http.StatusNonAuthoritativeInfo),
		http.StatusText(http.StatusNoContent),
		http.StatusText(http.StatusResetContent),
		http.StatusText(http.StatusPartialContent),
		http.StatusText(http.StatusMultiStatus),
		http.StatusText(http.StatusAlreadyReported),
		http.StatusText(http.StatusIMUsed),
		http.StatusText(http.StatusMultipleChoices),
		http.StatusText(http.StatusMovedPermanently),
		http.StatusText(http.StatusFound),
		http.StatusText(http.StatusSeeOther),
		http.StatusText(http.StatusNotModified),
		http.StatusText(http.StatusUseProxy),
		http.StatusText(http.StatusTemporaryRedirect),
		http.StatusText(http.StatusPermanentRedirect),
		http.StatusText(http.StatusBadRequest),
		http.StatusText(http.StatusUnauthorized),
		http.StatusText(http.StatusPaymentRequired),
		http.StatusText(http.StatusForbidden),
		http.StatusText(http.StatusNotFound),
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusText(http.StatusNotAcceptable),
		http.StatusText(http.StatusProxyAuthRequired),
		http.StatusText(http.StatusRequestTimeout),
		http.StatusText(http.StatusConflict),
		http.StatusText(http.StatusGone),
		http.StatusText(http.StatusLengthRequired),
		http.StatusText(http.StatusPreconditionFailed),
		http.StatusText(http.StatusRequestEntityTooLarge),
		http.StatusText(http.StatusRequestURITooLong),
		http.StatusText(http.StatusUnsupportedMediaType),
		http.StatusText(http.StatusRequestedRangeNotSatisfiable),
		http.StatusText(http.StatusExpectationFailed),
		http.StatusText(http.StatusTeapot),
		http.StatusText(http.StatusMisdirectedRequest),
		http.StatusText(http.StatusUnprocessableEntity),
		http.StatusText(http.StatusLocked),
		http.StatusText(http.StatusFailedDependency),
		http.StatusText(http.StatusTooEarly),
		http.StatusText(http.StatusUpgradeRequired),
		http.StatusText(http.StatusPreconditionRequired),
		http.StatusText(http.StatusTooManyRequests),
		http.StatusText(http.StatusRequestHeaderFieldsTooLarge),
		http.StatusText(http.StatusUnavailableForLegalReasons),
		http.StatusText(http.StatusInternalServerError),
		http.StatusText(http.StatusNotImplemented),
		http.StatusText(http.StatusBadGateway),
		http.StatusText(http.StatusServiceUnavailable),
		http.StatusText(http.StatusGatewayTimeout),
		http.StatusText(http.StatusHTTPVersionNotSupported),
		http.StatusText(http.StatusVariantAlsoNegotiates),
		http.StatusText(http.StatusInsufficientStorage),
		http.StatusText(http.StatusLoopDetected),
		http.StatusText(http.StatusNotExtended),
		http.StatusText(http.StatusNetworkAuthenticationRequired),
	}
	_span_status       = []string{STATUS_OK, STATUS_INFO, STATUS_WARN, STATUS_ERR, STATUS_CRITICAL}
	_all_priorities    = []int{PRIORITY_RULE_SAMPLER_REJECT, PRIORITY_USER_REJECT, PRIORITY_AUTO_REJECT, PRIORITY_AUTO_KEEP, PRIORITY_USER_KEEP, PRIORITY_RULE_SAMPLER_KEEP}
	_auto_priorities   = []int{PRIORITY_AUTO_KEEP, PRIORITY_AUTO_REJECT}
	_user_priorities   = []int{PRIORITY_USER_KEEP, PRIORITY_USER_REJECT}
	_smaple_priorities = []int{PRIORITY_RULE_SAMPLER_KEEP, PRIORITY_RULE_SAMPLER_REJECT}
)

func TestFindSpanTypeIntSpanID(t *testing.T) {
	trace := randDatakitTrace(t, 10, randService(_services...))
	parentialize(trace)
	parentIDs, spanIDs := gatherSpansInfo(trace)
	for i := range trace {
		switch FindSpanTypeStrSpanID(trace[i].SpanID, trace[i].ParentID, spanIDs, parentIDs) {
		case SPAN_TYPE_ENTRY:
			if i != 0 {
				t.Errorf("not an entry span")
				t.FailNow()
			}
		case SPAN_TYPE_LOCAL:
			if i == 0 || i == len(trace)-1 {
				t.Errorf("not one of local spans")
				t.FailNow()
			}
		case SPAN_TYPE_EXIT:
			if i != len(trace)-1 {
				t.Errorf("not an exit span")
				t.FailNow()
			}
		}
	}
}

func TestGetTraceInt64ID(t *testing.T) {
	for i := 0; i < 10; i++ {
		low := testutils.RandInt64(5)
		high := testutils.RandInt64(5)
		if fmt.Sprintf("%d", GetTraceInt64ID(high, low)) != strconv.Itoa(int(high))+strconv.Itoa(int(low)) {
			t.Error("get wrong trace id")
			t.FailNow()
		}
	}
}

func TestUnifyToInt64ID(t *testing.T) {
	ri := testutils.RandInt64StrID(10)
	testcases := map[string]int64{
		"345678987655678":                          345678987655678,
		"45f6f7f4d67a4b56":                         parseInt("45f6f7f4d67a4b56", 16, 64),
		"$%^&*&^%CGHGfxcghjsdkfh%^&6dr67d77855678": 3978710596982290232,
		"4%^&cvghjdfh":                             7167029555165947496,
		ri:                                         parseInt(ri, 10, 64),
	}
	for k, v := range testcases {
		if i := UnifyToInt64ID(k); i != v {
			t.Errorf("invalid transform origin: %s transform: %d expect: %d", k, i, v)
			t.FailNow()
		}
	}
}

func parseInt(s string, base int, bitSize int) int64 {
	i, _ := strconv.ParseInt(s, base, bitSize)

	return i
}

func parentialize(trace DatakitTrace) {
	if l := len(trace); l <= 1 {
		if l == 1 {
			trace[0].ParentID = "0"
		}

		return
	}

	trace[0].ParentID = "0"
	trace[0].SpanType = SPAN_TYPE_ENTRY
	for i := range trace[1:] {
		trace[i+1].TraceID = trace[0].TraceID
		trace[i+1].ParentID = trace[i].SpanID
	}
}

func gatherSpansInfo(trace DatakitTrace) (parentIDs, spanIDs map[string]bool) {
	parentIDs = make(map[string]bool)
	spanIDs = make(map[string]bool)
	for i := range trace {
		parentIDs[trace[i].ParentID] = true
		spanIDs[trace[i].SpanID] = true
	}

	return
}

func randDatakitTrace(t *testing.T, n int, opts ...randSpanOption) DatakitTrace {
	t.Helper()

	trace := make(DatakitTrace, n)
	for i := 0; i < n; i++ {
		trace[i] = randDatakitSpan(t, opts...)
	}

	return trace
}

func randDatakitSpan(t *testing.T, opts ...randSpanOption) *DatakitSpan {
	t.Helper()

	rand.Seed(time.Now().Local().UnixNano())
	dkspan := &DatakitSpan{
		TraceID:        testutils.RandStrID(30),
		ParentID:       testutils.RandStrID(30),
		SpanID:         testutils.RandStrID(30),
		Service:        testutils.RandString(30),
		Resource:       testutils.RandString(30),
		Operation:      testutils.RandString(30),
		Source:         testutils.RandString(30),
		SpanType:       testutils.RandString(10),
		SourceType:     testutils.RandString(10),
		Env:            testutils.RandString(100),
		Project:        testutils.RandString(10),
		Version:        testutils.RandVersion(10),
		EndPoint:       testutils.RandEndPoint(3),
		HTTPMethod:     testutils.RandString(10),
		HTTPStatusCode: testutils.RandString(10),
		ContainerHost:  testutils.RandString(20),
		PID:            testutils.RandInt64StrID(10),
		Start:          testutils.RandTime().Unix(),
		Duration:       testutils.RandInt64(5),
		Status:         testutils.RandString(10),
	}
	for i := range opts {
		opts[i](dkspan)
	}

	buf, err := json.Marshal(dkspan)
	if err != nil {
		t.Error(err.Error())
	}
	dkspan.Content = string(buf)

	return dkspan
}

type randSpanOption func(dkspan *DatakitSpan)

func randService(services ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Service = testutils.RandWithinStrings(services)
		}
	}
}

func randResource(resources ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Resource = testutils.RandWithinStrings(resources)
		}
	}
}

func randSource(sources ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Source = testutils.RandWithinStrings(sources)
		}
	}
}

func randSpanTypes(types ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.SpanType = testutils.RandWithinStrings(types)
		}
	}
}

func randHTTPMethod(methods ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.HTTPMethod = testutils.RandWithinStrings(methods)
		}
	}
}

func randHTTPStatusCode(codes ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.HTTPStatusCode = testutils.RandWithinStrings(codes)
		}
	}
}

func randSpanStatus(status ...string) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Status = testutils.RandWithinStrings(status)
		}
	}
}

func randPriority(priorities ...int) randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Metrics = make(map[string]interface{})
			dkspan.Metrics[FIELD_PRIORITY] = testutils.RandWithinInts(priorities)
		}
	}
}

func randTags() randSpanOption {
	return func(dkspan *DatakitSpan) {
		if dkspan != nil {
			dkspan.Tags = testutils.RandTags(10, 10, 20)
		}
	}
}
