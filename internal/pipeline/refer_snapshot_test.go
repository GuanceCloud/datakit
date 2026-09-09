// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

// Exercise the execution adapter, not merely the identity of captured pointers.
// Instances may still refresh their contents; this is not a frozen data snapshot.
func TestReferSnapshotScriptExecution(t *testing.T) {
	originalManager, _ := plval.GetManager()
	originalTable, _ := plval.GetRefTb()
	t.Cleanup(func() { plval.SetManager(originalManager); plval.SetRefTb(originalTable) })
	plval.SetManager(plval.NewScriptManager(nil, nil))

	newTable := func(label string) *refertable.ReferTable {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `[{"table_name":"shared","column_name":["id","label"],"column_type":["int","string"],"row_data":[[7,%q]]}]`, label)
		}))
		t.Cleanup(server.Close)
		table, err := refertable.NewReferTable(refertable.RefTbCfg{URL: server.URL, Interval: time.Second})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { defer close(done); table.PullWorker(ctx) }()
		t.Cleanup(func() {
			cancel()
			select {
			case <-done:
				if closer, ok := any(table).(interface{ Close() error }); ok {
					if err := closer.Close(); err != nil {
						t.Error(err)
					}
				}
			case <-time.After(2 * time.Second):
				t.Error("refer worker failed to stop")
			}
		})
		if !table.InitFinished(2 * time.Second) {
			t.Fatal("refer initialization timed out")
		}
		return table
	}
	old, next := newTable("old"), newTable("new")
	capture := func(table *refertable.ReferTable) refertable.PlReferTables {
		t.Helper()
		plval.SetRefTb(table)
		manager, runner, offload, snapshot, _, err := acquirePipelineLeasesContext(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(manager.Release)
		if runner != nil {
			t.Cleanup(runner.release)
		}
		if offload != nil {
			t.Cleanup(offload.Release)
		}
		return snapshot
	}
	emptySnapshot := capture(nil)
	oldSnapshot := capture(old)
	newSnapshot := capture(next)
	// All executions happen with next published, including the old/empty batches.
	for _, source := range []string{
		`query_refer_table("shared","id",7); add_key(after,true)`,
		`mquery_refer_table("shared",["id"],[7]); add_key(after,true)`,
	} {
		script, err := NewPlScriptSimple(point.Logging, "refer-snapshot.p", source)
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			name     string
			snapshot refertable.PlReferTables
			want     any
		}{
			{"old_batch", oldSnapshot, "old"},
			{"new_batch", newSnapshot, "new"},
			{"absent_at_batch_start", emptySnapshot, nil},
		} {
			t.Run(test.name, func(t *testing.T) {
				for i := 0; i < 10; i++ {
					pt := point.NewPoint("refer", point.NewKVs(map[string]any{"sequence": int64(i)}))
					run := pointRun{point: pt, script: script, referCaptured: true, referSnapshot: test.snapshot}
					runPipelineGoContext(context.Background(), point.Logging, &run, nil)
					if run.output == nil || run.dropped || len(run.created) != 0 {
						t.Fatalf("unexpected terminal outcome: %+v", run)
					}
					if got := run.output.Get("label"); got != test.want {
						t.Fatalf("record %d label=%v want=%v", i, got, test.want)
					}
					if run.output.Get("after") != true || run.output.Get("sequence") != int64(i) {
						t.Fatalf("record %d did not preserve input and execute following statement", i)
					}
				}
			})
		}
	}
}
