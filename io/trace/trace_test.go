package trace

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func randDatakitSpan(t *testing.T) *DatakitSpan {
	t.Helper()

	dkspan := &DatakitSpan{
		TraceID:            testutils.RandStrID(30),
		ParentID:           testutils.RandStrID(30),
		SpanID:             testutils.RandStrID(30),
		Service:            testutils.RandString(30),
		Resource:           testutils.RandString(30),
		Operation:          testutils.RandString(30),
		Source:             testutils.RandString(30),
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
