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

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

var (
	_services     = []string{"login", "game", "fire_gun", "march", "kill", "logout"}
	_resources    = []string{"/get_user/name", "/push/data", "/check/security", "/fetch/data_source", "/pull/all_data", "/list/user_name"}
	_source       = []string{"ddtrace", "jaeger", "opentelemetry", "skywalking", "zipkin"}
	_span_types   = []string{SpanTypeEntry, SpanTypeLocal, SpanTypeExit, SpanTypeUnknown}
	_source_types = []string{SpanSourceApp, SpanSourceCache, SpanSourceCustomer, SpanSourceDb, SpanSourceWeb}
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
	_span_status       = []string{StatusOk, StatusInfo, StatusWarn, StatusErr, StatusCritical}
	_all_priorities    = []int{PriorityRuleSamplerReject, PriorityUserReject, PriorityAutoReject, PriorityAutoKeep, PriorityUserKeep, PriorityRuleSamplerKeep}
	_auto_priorities   = []int{PriorityAutoKeep, PriorityAutoReject}
	_user_priorities   = []int{PriorityUserKeep, PriorityUserReject}
	_smaple_priorities = []int{PriorityRuleSamplerKeep, PriorityRuleSamplerReject}
)

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

func parseInt(s string, base int, bitSize int) uint64 {
	i, _ := strconv.ParseUint(s, base, bitSize)

	return i
}

func parentialize(trace DatakitTrace) {
	if l := len(trace); l <= 1 {
		if l == 1 {
			trace[0].MustAdd(FieldParentID, "0")
		}

		return
	}

	trace[0].MustAdd(FieldParentID, "0")
	trace[0].MustAdd(TagSpanType, SpanTypeEntry)
	for i := range trace[1:] {
		trace[i+1].MustAdd(FieldTraceID, trace[0].GetFiledToString(FieldTraceID))
		trace[i+1].MustAdd(FieldParentID, trace[i].GetFiledToString(FieldSpanid))
	}
}

/*
	func gatherSpansInfo(trace DatakitTrace) (parentIDs, spanIDs map[string]bool) {
		parentIDs = make(map[string]bool)
		spanIDs = make(map[string]bool)
		for i := range trace {
			parentIDs[trace[i].ParentID] = true
			spanIDs[trace[i].SpanID] = true
		}

		return
	}
*/
func randDatakitTrace(t *testing.T, n int, opts ...randSpanOption) DatakitTrace {
	t.Helper()

	trace := make(DatakitTrace, n)
	for i := 0; i < n; i++ {
		trace[i] = randDatakitSpan(t, opts...)
	}

	return trace
}

func randDatakitSpan(t *testing.T, opts ...randSpanOption) *DkSpan {
	t.Helper()

	rand.Seed(time.Now().Local().UnixNano())
	var spanKV point.KVs
	spanKV = spanKV.Add(FieldTraceID, testutils.RandStrID(30), false, false).
		Add(FieldParentID, testutils.RandStrID(30), false, false).
		Add(FieldSpanid, testutils.RandStrID(30), false, false).
		Add(TagService, testutils.RandString(30), true, false).
		Add(FieldResource, testutils.RandString(30), false, false).
		Add(TagOperation, testutils.RandString(30), true, false).
		Add(TagSource, testutils.RandString(30), true, false).
		Add(TagSpanType, testutils.RandString(10), true, false).
		Add(TagSourceType, testutils.RandString(10), true, false).
		Add(FieldStart, testutils.RandTime().Unix(), false, false).
		Add(FieldDuration, testutils.RandInt64(5), false, false).
		Add(TagSpanStatus, testutils.RandString(10), true, false).
		AddTag(TagProject, testutils.RandString(10)).
		AddTag(TagVersion, testutils.RandVersion(10)).
		AddTag(TagEndpoint, testutils.RandEndPoint(3)).
		AddTag(TagPid, testutils.RandInt64StrID(10)).
		AddTag(TagContainerHost, testutils.RandString(20))

	pt := point.NewPointV2("testSource", spanKV, point.DefaultLoggingOptions()...)
	dkSpan := &DkSpan{pt}
	for i := range opts {
		opts[i](dkSpan)
	}

	buf, err := json.Marshal(dkSpan)
	if err != nil {
		t.Error(err.Error())
	}
	// dkspan.Content = string(buf)
	dkSpan.MustAdd(FieldMessage, string(buf))

	return dkSpan
}

type randSpanOption func(dkspan *DkSpan)

func randService(services ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			dkspan.MustAddTag(TagService, testutils.RandWithinStrings(services))
		}
	}
}

func randResource(resources ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			// dkspan.Resource = testutils.RandWithinStrings(resources)
			dkspan.MustAdd(FieldResource, testutils.RandWithinStrings(resources))
		}
	}
}

func randSource(sources ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			dkspan.MustAdd(TagSource, testutils.RandWithinStrings(sources))
			// dkspan.Source = testutils.RandWithinStrings(sources)
		}
	}
}

func randSpanTypes(types ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			// dkspan.SpanType = testutils.RandWithinStrings(types)
			dkspan.MustAdd(TagSpanType, testutils.RandWithinStrings(types))
		}
	}
}

func randHTTPMethod(methods ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		dkspan.MustAddTag(TagHttpMethod, testutils.RandWithinStrings(methods))
	}
}

func randHTTPStatusCode(codes ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		dkspan.MustAddTag(TagHttpStatusCode, testutils.RandWithinStrings(codes))
	}
}

func randSpanStatus(status ...string) randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			dkspan.MustAdd(TagSpanStatus, testutils.RandWithinStrings(status))
		}
	}
}

func randPriority(priorities ...int) randSpanOption {
	return func(dkspan *DkSpan) {
		dkspan.MustAdd(FieldPriority, testutils.RandWithinInts(priorities))
	}
}

func randTags() randSpanOption {
	return func(dkspan *DkSpan) {
		if dkspan != nil {
			rTags := testutils.RandTags(10, 10, 20)
			for k, v := range rTags {
				dkspan.MustAddTag(k, v)
			}
		}
	}
}
