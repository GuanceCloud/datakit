// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jsonfields

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		converted      bool
		expectedFields map[string]any
	}{
		{
			name:      "all supported types",
			input:     `{"string":"value","integer":42,"negative":-1,"decimal":1.5,"exponent":1e3,"bool":true,"object":{"b":2,"a":1},"array":[1,"two"],"null":null}`,
			converted: true,
			expectedFields: map[string]any{
				"string":   "value",
				"integer":  int64(42),
				"negative": int64(-1),
				"decimal":  float64(1.5),
				"exponent": float64(1000),
				"bool":     true,
				"object":   `{"a":1,"b":2}`,
				"array":    `[1,"two"]`,
			},
		},
		{
			name:      "reserved fields use deterministic suffixes",
			input:     `{"time":1,"json_time":2,"json_time_1":3,"source":"src","json_source":"existing","date":true,"storage_index":"idx","json_storage_index":"existing-index"}`,
			converted: true,
			expectedFields: map[string]any{
				"json_time":            int64(2),
				"json_time_1":          int64(3),
				"json_time_2":          int64(1),
				"json_source":          "existing",
				"json_source_1":        "src",
				"json_date":            true,
				"json_storage_index":   "existing-index",
				"json_storage_index_1": "idx",
			},
		},
		{
			name:      "storage index uses reserved prefix",
			input:     `{"storage_index":"custom"}`,
			converted: true,
			expectedFields: map[string]any{
				"json_storage_index": "custom",
			},
		},
		{
			name:      "empty storage index uses reserved prefix",
			input:     `{"storage_index":""}`,
			converted: true,
			expectedFields: map[string]any{
				"json_storage_index": "",
			},
		},
		{
			name:      "non-string storage index uses reserved prefix",
			input:     `{"storage_index":123}`,
			converted: true,
			expectedFields: map[string]any{
				"json_storage_index": int64(123),
			},
		},
		{
			name:           "integer overflow is ignored",
			input:          `{"overflow":9223372036854775808,"valid":1}`,
			converted:      true,
			expectedFields: map[string]any{"valid": int64(1)},
		},
		{
			name:           "float overflow is ignored",
			input:          `{"overflow":1e400,"valid":1.5}`,
			converted:      true,
			expectedFields: map[string]any{"valid": float64(1.5)},
		},
		{
			name:           "JSON message is retained",
			input:          `{"message":"hello"}`,
			converted:      true,
			expectedFields: map[string]any{"message": "hello"},
		},
		{
			name:      "only unusable fields falls back",
			input:     `{"overflow":9223372036854775808,"null":null}`,
			converted: false,
		},
		{name: "empty object falls back", input: `{}`, converted: false},
		{name: "invalid JSON falls back", input: `{"key":`, converted: false},
		{name: "trailing JSON falls back", input: `{} {}`, converted: false},
		{name: "array root falls back", input: `[1,2]`, converted: false},
		{name: "scalar root falls back", input: `true`, converted: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Convert([]byte(tc.input))
			assert.Equal(t, tc.converted, result.Converted)
			assert.Equal(t, tc.expectedFields, result.Fields)
		})
	}
}

func TestConversionApplyReplacesTagWithField(t *testing.T) {
	conversion := Convert([]byte(`{"host":"json-host","status":1}`))
	require.True(t, conversion.Converted)

	kvs := point.KVs{}.
		AddTag("host", "collector-host").
		AddTag("status", "unknown")
	kvs = conversion.Apply(kvs)

	pt := point.NewPoint("logging", kvs, point.WithPrecheck(false))
	assert.Empty(t, pt.GetTag("host"))
	assert.Equal(t, "json-host", pt.Get("host"))
	assert.Empty(t, pt.GetTag("status"))
	assert.Equal(t, int64(1), pt.Get("status"))
	assert.Len(t, pt.KVs(), 2)
}

func TestConversionApplyTrimsFinalFieldsDeterministically(t *testing.T) {
	fields := make(map[string]any, 1025)
	for i := 0; i < 1024; i++ {
		fields[fmt.Sprintf("json_%04d", i)] = int64(i)
	}
	fields["zz_tag"] = "json-field"

	kvs := point.KVs{}.
		AddTag("zz_tag", "collector-tag").
		Add("collector_first", "kept").
		AddTag("host", "collector-host").
		Add("collector_second", "kept")
	kvs = (Conversion{Fields: fields, Converted: true}).Apply(kvs)

	assert.Equal(t, 1024, kvs.FieldCount())
	assert.Equal(t, 1, kvs.TagCount())
	require.NotNil(t, kvs.Get("collector_first"))
	require.NotNil(t, kvs.Get("collector_second"))
	require.NotNil(t, kvs.Get("json_1021"))
	assert.Nil(t, kvs.Get("json_1022"))
	assert.Nil(t, kvs.Get("json_1023"))
	assert.Nil(t, kvs.Get("zz_tag"))
}

func TestConvertNormalizesPointKeys(t *testing.T) {
	tests := []struct {
		name      string
		root      map[string]any
		converted bool
		fields    map[string]any
	}{
		{
			name:      "unchanged key wins normalized collision",
			root:      map[string]any{"a.b": 1, "a_b": 2},
			converted: true,
			fields:    map[string]any{"a_b": int64(2)},
		},
		{
			name:      "lexical key wins normalized collision",
			root:      map[string]any{"a..b": 1, "a._b": 2},
			converted: true,
			fields:    map[string]any{"a__b": int64(1)},
		},
		{
			name: "point key adjustments",
			root: map[string]any{
				"line\nkey": 1,
				`tail\\`:    2,
			},
			converted: true,
			fields: map[string]any{
				"line key": int64(1),
				"tail":     int64(2),
			},
		},
		{
			name:      "long key is truncated",
			root:      map[string]any{strings.Repeat("k", 257): 1},
			converted: true,
			fields:    map[string]any{strings.Repeat("k", 256): int64(1)},
		},
		{
			name: "empty normalized keys fall back",
			root: map[string]any{
				"":     1,
				`\\\\`: 2,
			},
			converted: false,
		},
		{
			name: "normalized reserved keys use reserved names",
			root: map[string]any{
				"source":         "exact-source",
				`source\`:        "normalized-source",
				`date\`:          "normalized-date",
				`time\`:          123,
				`storage_index\`: "normalized-index",
			},
			converted: true,
			fields: map[string]any{
				"json_source":        "exact-source",
				"json_date":          "normalized-date",
				"json_time":          int64(123),
				"json_storage_index": "normalized-index",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.root)
			require.NoError(t, err)

			conversion := Convert(data)
			assert.Equal(t, tc.converted, conversion.Converted)
			assert.Equal(t, tc.fields, conversion.Fields)
		})
	}
}
