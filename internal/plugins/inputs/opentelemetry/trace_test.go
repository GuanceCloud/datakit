// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/hex"
	"testing"
	T "testing"

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
			assert.NotNil(t, pt.Get(itrace.TagOperation))
			assert.NotNil(t, pt.Get(itrace.TagSource))
			assert.NotNil(t, pt.Get(itrace.TagService))

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
