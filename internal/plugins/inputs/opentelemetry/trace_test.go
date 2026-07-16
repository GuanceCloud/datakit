// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/hex"
	"testing"
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	cv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	"github.com/stretchr/testify/assert"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestConvert(t *T.T) {
	id, _ := hex.DecodeString("818616084f850520843d19e3936e4720")
	t.Logf("id len=%d", len(id))
	type args struct {
		id []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "case128", args: args{id: id}, want: "9528800851807586080"},
	}

	ipt := defaultInput()
	ipt.CompatibleDDTrace = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipt.convertBinID(tt.args.id); got != tt.want {
				t.Errorf("convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertZeroParentID(t *T.T) {
	zeroParentID := make([]byte, 8)
	assert.Equal(t, "0000000000000000", hex.EncodeToString(zeroParentID))

	assert.Equal(t, "0", byteToString(make([]byte, 8)))
}

func Test_getSourceType(t *T.T) {
	var kvs point.KVs
	kvs = kvs.Add(otelHTTPMethodKey, "GET")
	kvs = kvs.Add(otelDBSystemKey, "postgresql")

	assert.Equal(t, itrace.SpanSourceDb, getSourceType(kvs))

	kvs = point.KVs{}.Add(otelMessagingDestKey, "topic")
	assert.Equal(t, itrace.SpanSourceMsgque, getSourceType(kvs))

	kvs = point.KVs{}.Add(otelRPCGRPCStatusKey, "0")
	assert.Equal(t, itrace.SpanSourceWeb, getSourceType(kvs))
}

func Test_traceBaseService(t *T.T) {
	ipt := defaultInput()

	for _, tt := range []struct {
		name  string
		attrs map[string]string
		want  string
	}{
		{
			name: "gRPC keeps original service name",
			attrs: map[string]string{
				"rpc.system":       "grpc",
				"messaging.system": "kafka",
			},
		},
		{
			name: "gRPC semantic convention name keeps original service name",
			attrs: map[string]string{
				"rpc.system.name": "grpc",
			},
		},
		{
			name: "non-gRPC RPC splits service name",
			attrs: map[string]string{
				"rpc.system": "thrift",
			},
			want: "thrift",
		},
	} {
		t.Run(tt.name, func(t *T.T) {
			assert.Equal(t, tt.want, ipt.traceBaseService(tt.attrs))
		})
	}
}

func Test_commonTagFields(t *T.T) {
	t.Run("tag-fields", func(t *T.T) {
		nspans := 10
		traces := createTestTraceData(nspans)

		ipt := defaultInput()
		ipt.setup()
		ipt.Tags = map[string]string{
			"foo": "bar",
		}

		arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

		assert.Len(t, arr, 1)         // only 1 trace
		assert.Len(t, arr[0], nspans) // and n spans in that trace

		traceID := arr[0][0].Get(itrace.FieldTraceID) // all span's trace id are the same

		for _, pt := range arr[0] {
			assert.NotNil(t, traceID, pt.Get(itrace.FieldTraceID)) // all span get same trace ID
			assert.NotNil(t, pt.Get(itrace.FieldParentID))
			assert.NotNil(t, pt.Get(itrace.FieldSpanid))
			assert.NotNil(t, pt.Get(itrace.FieldResource))
			assert.NotNil(t, pt.Get(itrace.FieldStart))
			assert.NotNil(t, pt.Get(itrace.FieldDuration))
			assert.NotNil(t, pt.Get(itrace.TagSpanStatus))
			assert.NotNil(t, pt.Get(itrace.TagDKFingerprintKey))
			assert.NotNil(t, pt.Get(itrace.TagSpanType))
			assert.NotNil(t, pt.Get(itrace.TagSource))
			assert.NotNil(t, pt.Get(itrace.TagService))
			assert.Equal(t, inputName, pt.Name())
			assert.Equal(t, inputName, pt.Get(itrace.TagSource))
			assert.Equal(t, "db", pt.Get(itrace.TagSourceType))
			assert.Contains(t, []string{itrace.SpanTypeEntry, itrace.SpanTypeExit, itrace.SpanTypeLocal}, pt.Get(itrace.TagSpanType))
			assert.Equal(t, "localhost", pt.Get(itrace.TagCollectorSourceIP))

			// SplitServiceName()
			assert.Equal(t, "postgresql", pt.Get(itrace.TagService))
			assert.Equal(t, "test-service", pt.Get(itrace.TagBaseService))
			assert.Equal(t, "mocked-runtime-id", pt.Get(itrace.FieldRuntimeID))

			// input tags
			assert.Equal(t, "bar", pt.Get("foo"))

			// exception events
			assert.Equal(t, "mocked-exception-type", pt.Get(itrace.FieldErrType))
			assert.Equal(t, "mocked-exception-message", pt.Get(itrace.FieldErrMessage))
			assert.Equal(t, "mocked-exception-stack", pt.Get(itrace.FieldErrStack))

			// span kind
			assert.Equal(t, "server", pt.Get(itrace.TagSpanKind))

			// merged attrs
			assert.Contains(t, pt.Get("message"), "scope-double-attr")
			assert.Contains(t, pt.Get("message"), "scope-byte-attr")
			assert.Contains(t, pt.Get("message"), "scope-int-attr")
			assert.Contains(t, pt.Get("message"), "scope-str-attr")
			assert.Contains(t, pt.Get("message"), "scope-list-attr")
			assert.Contains(t, pt.Get("message"), "scope-kv-attr")

			t.Logf("%s", pt.Pretty())
		}
	})
}

type staticGlobalTagger struct {
	hostTags map[string]string
}

func (g *staticGlobalTagger) HostTags() map[string]string     { return g.hostTags }
func (g *staticGlobalTagger) ElectionTags() map[string]string { return nil }
func (g *staticGlobalTagger) Updated() bool                   { return false }
func (g *staticGlobalTagger) UpdateVersion()                  {}

func Test_parseResourceSpansGlobalTags(t *T.T) {
	traces := createTestTraceData(1)

	t.Run("default", func(t *T.T) {
		ipt := defaultInput()
		ipt.setup()
		ipt.Tagger = &staticGlobalTagger{hostTags: map[string]string{"global_key": "global_value"}}

		arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

		assert.Len(t, arr, 1)
		assert.Equal(t, "global_value", arr[0][0].GetTag("global_key"))
	})

	t.Run("tracing metric switch keeps global tags", func(t *T.T) {
		ipt := defaultInput()
		ipt.setup()
		ipt.TracingMetricDisableGlobalHostTags = true
		ipt.Tagger = &staticGlobalTagger{hostTags: map[string]string{"global_key": "global_value"}}

		arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

		assert.Len(t, arr, 1)
		assert.Equal(t, "global_value", arr[0][0].GetTag("global_key"))
	})
}

func Test_parseResourceSpansLoongSuiteDBSystemName(t *T.T) {
	traces := createTestTraceData(1)
	span := traces.ResourceSpans[0].ScopeSpans[0].Spans[0]
	span.Attributes = []*cv1.KeyValue{
		{
			Key: "db.system.name",
			Value: &cv1.AnyValue{
				Value: &cv1.AnyValue_StringValue{StringValue: "mysql"},
			},
		},
		{
			Key: "db.operation.name",
			Value: &cv1.AnyValue{
				Value: &cv1.AnyValue_StringValue{StringValue: "SELECT"},
			},
		},
		{
			Key: "db.query.text",
			Value: &cv1.AnyValue{
				Value: &cv1.AnyValue_StringValue{StringValue: "select * from t"},
			},
		},
		{
			Key: "server.address",
			Value: &cv1.AnyValue{
				Value: &cv1.AnyValue_StringValue{StringValue: "mysql.local"},
			},
		},
	}

	ipt := defaultInput()
	ipt.setup()
	arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

	assert.Len(t, arr, 1)
	assert.Len(t, arr[0], 1)
	assert.Equal(t, "mysql", arr[0][0].Get("db_system"))
	assert.Equal(t, "SELECT", arr[0][0].Get("db_operation"))
	assert.Equal(t, "select * from t", arr[0][0].Get("db_statement"))
	assert.Equal(t, "mysql", arr[0][0].Get(itrace.TagService))
	assert.Equal(t, "test-service", arr[0][0].Get(itrace.TagBaseService))
	assert.Equal(t, "mysql.local", arr[0][0].Get(itrace.TagDBHost))
	assert.Equal(t, itrace.SpanSourceDb, arr[0][0].Get(itrace.TagSourceType))
}

func Test_parseResourceSpansMessagingDestinationName(t *T.T) {
	traces := createTestTraceData(1)
	span := traces.ResourceSpans[0].ScopeSpans[0].Spans[0]
	span.Attributes = []*cv1.KeyValue{
		{
			Key: "messaging.destination.name",
			Value: &cv1.AnyValue{
				Value: &cv1.AnyValue_StringValue{StringValue: "orders"},
			},
		},
	}

	ipt := defaultInput()
	ipt.setup()
	arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

	assert.Len(t, arr, 1)
	assert.Len(t, arr[0], 1)
	assert.Equal(t, itrace.SpanSourceMsgque, arr[0][0].Get(itrace.TagSourceType))
	assert.Equal(t, "orders", arr[0][0].Get(otelMessagingDestKey))
	assert.Nil(t, arr[0][0].Get("messaging_destination.name"))
}

func Test_customTags(t *T.T) {
	t.Run("custome_tags", func(t *T.T) {
		traces := createTestTraceData(1)

		ipt := defaultInput()
		ipt.CustomerTags = []string{"project.id", "app_id"}

		ipt.setup()

		arr := ipt.parseResourceSpans(traces.ResourceSpans, "localhost")

		assert.Len(t, arr, 1)
		assert.Equal(t, "project-001", arr[0][0].Get("project_id"))
		assert.Equal(t, "app-001", arr[0][0].Get("app_id"))

		assert.NotContains(t, arr[0][0].Get("message"), "project.id") // custome tags should remove after extracted
		assert.Contains(t, arr[0][0].Get("message"), "not-used-value")

		t.Logf("%s", arr[0][0].Pretty())
	})
}
