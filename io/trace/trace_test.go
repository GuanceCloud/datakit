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

func TestFindSpanTypeIntSpanID(t *testing.T) {
	trace := randDatakitTrace(t, 10)
	parentialize(trace)
	parentIDs, spanIDs := extractTraceIDs(trace)
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
	for i := range trace[1:] {
		trace[i+1].TraceID = trace[0].TraceID
		trace[i+1].ParentID = trace[i].SpanID
	}
}

func extractTraceIDs(trace DatakitTrace) (parentids, spanids map[string]bool) {
	parentids = make(map[string]bool)
	spanids = make(map[string]bool)
	for i := range trace {
		parentids[trace[i].ParentID] = true
		spanids[trace[i].SpanID] = true
	}

	return
}

func randDatakitTraceByService(t *testing.T, n int, service, resource, source string) DatakitTrace {
	t.Helper()

	trace := randDatakitTrace(t, n)
	for i := range trace {
		trace[i].Service = service
		trace[i].Resource = resource
		trace[i].Source = source
	}

	return trace
}

func randDatakitTrace(t *testing.T, n int) DatakitTrace {
	t.Helper()

	trace := make(DatakitTrace, n)
	for i := 0; i < n; i++ {
		trace[i] = randDatakitSpan(t)
	}

	return trace
}

func randDatakitSpan(t *testing.T) *DatakitSpan {
	t.Helper()

	rand.Seed(time.Now().Local().UnixNano())
	dkspan := &DatakitSpan{
		TraceID:            testutils.RandStrID(30),
		ParentID:           testutils.RandStrID(30),
		SpanID:             testutils.RandStrID(30),
		Service:            testutils.RandString(30),
		Resource:           testutils.RandString(30),
		Operation:          testutils.RandString(30),
		Source:             testutils.RandWithinStrings([]string{"ddtrace", "jaeger", "skywalking", "zipkin"}),
		SpanType:           testutils.RandWithinStrings([]string{SPAN_TYPE_ENTRY, SPAN_TYPE_LOCAL, SPAN_TYPE_EXIT, SPAN_TYPE_UNKNOW}),
		SourceType:         testutils.RandWithinStrings([]string{SPAN_SERVICE_APP, SPAN_SERVICE_CACHE, SPAN_SERVICE_CUSTOM, SPAN_SERVICE_DB, SPAN_SERVICE_WEB}),
		Env:                testutils.RandString(100),
		Project:            testutils.RandString(10),
		Version:            testutils.RandVersion(10),
		Tags:               testutils.RandTags(10, 10, 20),
		EndPoint:           testutils.RandEndPoint(3),
		HTTPMethod:         testutils.RandWithinStrings([]string{http.MethodPost, http.MethodGet}),
		HTTPStatusCode:     testutils.RandWithinStrings([]string{"200", "400", "403", "404", "500"}),
		ContainerHost:      testutils.RandString(20),
		PID:                testutils.RandInt64StrID(10),
		Start:              testutils.RandTime().Unix(),
		Duration:           testutils.RandInt64(5),
		Status:             testutils.RandWithinStrings([]string{STATUS_OK, STATUS_INFO, STATUS_WARN, STATUS_ERR, STATUS_CRITICAL}),
		Priority:           testutils.RandWithinInts([]int{PriorityReject, PriorityAuto, PriorityAuto}),
		SamplingRateGlobal: rand.Float64(),
	}
	buf, err := json.Marshal(dkspan)
	if err != nil {
		t.Error(err.Error())
	}
	dkspan.Content = string(buf)

	return dkspan
}
