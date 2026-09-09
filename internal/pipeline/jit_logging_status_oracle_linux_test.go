// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
)

// Real logging collectors supply status as a tag. Verify physical key kinds,
// not only values, against the unmodified Go runner for both output protocols.
func TestJITLoggingStatusKindOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{name: "preserve", source: `add_key(result, 1)`},
		{name: "json", source: `json(_, status); json(_, value)`, input: `{"status":"ERROR","value":42}`},
		{name: "grok", source: `grok(_, "%{WORD:method} %{INT:code}"); cast(code,"int")`, input: "GET 200"},
		{name: "grok_status", source: `grok(_, "%{WORD:status}")`, input: "WARN"},
		{name: "multiline", source: `grok(_, "%{LOGLEVEL:level} %{GREEDYDATA:detail}")`, input: "ERROR failed\n  at worker.run"},
		{name: "dynamic", source: `data=load_json(message); pt_kvs_set(data["key"],data["value"])`, input: `{"key":"status","value":"WARN"}`},
		{name: "disabled", source: `setopt(status_mapping=false); add_key(result,1)`},
		{name: "delete", source: `drop_key(status)`},
		{name: "default_time", source: `json(_, time); default_time(time,"UTC")`, input: `{"time":"2026-09-09 10:00:00"}`},
		{name: "default_time_invalid", source: `json(_, time); default_time(time,"UTC")`, input: `{"time":"invalid"}`},
		{name: "default_time_with_fmt", source: `json(_, time); default_time_with_fmt(time,"2006-01-02 15:04:05","UTC")`, input: `{"time":"2026-09-09 10:00:00"}`},
		{name: "prefix_error", source: `add_key(status,"WARN"); add_key(result,1/divisor)`, fails: true},
	}
	for _, kind := range []string{"tag", "field", "missing"} {
		t.Run(kind, func(t *testing.T) {
			for _, status := range []string{"info", "WARN", "CUSTOM", ""} {
				t.Run("status="+status, func(t *testing.T) {
					runJSONOracle(t, cases, func(pt *point.Point) {
						switch kind {
						case "tag":
							pt.SetTag("status", status)
						case "field":
							pt.Set("status", status)
						default:
							pt.Del("status")
						}
					})
				})
			}
		})
	}
}
