// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

func TestReferTablePipelineGoUpstreamCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	table := newUpstreamReferTable(t)
	runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 16,
		ServiceConfig{Version: 1}, func() (*HostCompat, error) {
			return NewPipelineGoHostWithReferObserved(nil, table.Tables(), nil), nil
		})
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name, source string
		expected     map[string]any
	}{
		{"query-named", `add_key(f1,123); query_refer_table("table1",key="f1",value=f1)`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"mquery-named", `add_key(f1,123); mquery_refer_table("table1",["f1"],values=[f1])`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"int", `query_refer_table("table1","f1",123)`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"string", `query_refer_table("table1","key1","a")`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"dynamic-table-key-value", `t="table1"; k="key1"; v="a"; query_refer_table(t,k,v)`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"float", `query_refer_table("table1","key2",123.)`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"bool", `query_refer_table("table1","f2",false)`, upstreamReferRow("a", float64(123), int64(123), false)},
		{"float-not-int", `query_refer_table("table1","key2",123)`, nil},
		{"missing-point-key", `query_refer_table("table1","f1",f1)`, nil},
		{"named-value-with-positional-key", `query_refer_table("table1","f1",value=f1)`, nil},
		{"multi-dynamic", `key="f2"; value="ab"; mquery_refer_table("table1",["key1",key],[value,false])`, upstreamReferRow("ab", float64(1234), int64(123), false)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := runner.Check(tc.source)
			if check.Route != RouteJITWithHost || check.Capabilities.Backend != ExecutionBackendMachineCode ||
				check.Capabilities.RequiredHostFlags == 0 || len(check.Capabilities.HostCalls) != 1 || check.Capabilities.HostCalls[0] != "refer_table" {
				t.Fatalf("route=%+v", check)
			}
			for _, size := range []int{1, 2, 4, 8, 10, 128} {
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"message": tc.name, "sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("batch %d records=%d error=%v", size, len(batch.Records), err)
				}
				for i := range batch.Records {
					if batch.Records[i].Status != TerminalOK {
						t.Fatalf("batch %d record %d=%+v", size, i, batch.Records[i])
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					for _, key := range []string{"key1", "key2", "f1", "f2"} {
						want, present := tc.expected[key]
						got, exists := points[i].Fields[key]
						if present != exists || (present && !referValueEqual(got, want)) {
							t.Fatalf("batch %d record %d key %s=%#v/%v want=%#v/%v: %#v", size, i, key, got, exists, want, present, points[i])
						}
					}
					if points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("batch %d record %d lost input: %#v", size, i, points[i])
					}
				}
			}
		})
	}
}

func upstreamReferRow(key1 string, key2 float64, f1 int64, f2 bool) map[string]any {
	return map[string]any{"key1": key1, "key2": key2, "f1": f1, "f2": f2}
}

func referValueEqual(got, want any) bool {
	// JSON host responses normalize integral JSON numbers to int64 at the DataKit
	// ABI boundary while the upstream in-process table exposes float64. Preserve
	// strict type comparisons except for this wire-normalized numeric equivalent.
	if left, ok := got.(int64); ok {
		if right, ok := want.(float64); ok {
			return float64(left) == right
		}
	}
	return fmt.Sprint(got) == fmt.Sprint(want) && fmt.Sprintf("%T", got) == fmt.Sprintf("%T", want)
}

func newUpstreamReferTable(t *testing.T) *refertable.ReferTable {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, upstreamReferTableData)
	}))
	t.Cleanup(server.Close)
	table, err := refertable.NewReferTable(refertable.RefTbCfg{URL: server.URL, Interval: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		table.PullWorker(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("refer worker did not stop")
		}
		if closer, ok := any(table).(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	if !table.InitFinished(3 * time.Second) {
		t.Fatal("refer table initialization timed out")
	}
	return table
}

const upstreamReferTableData = `[
  {"table_name":"table1","column_name":["key1","key2","f1","f2"],
   "column_type":["string","float","int","bool"],
   "row_data":[["a",123,"123","false"],["ab","1234","123","true"],["ab","1234","123","false"]]},
  {"table_name":"table2","primary_key":["key1","key2"],
   "column_name":["key1","key2","f1","f2"],"column_type":["string","float","int","bool"],
   "row_data":[["a",123,"123","true"],["a","1234","123","true"]]}
]`
