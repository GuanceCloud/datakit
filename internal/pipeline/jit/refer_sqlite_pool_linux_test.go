// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

const validSQLiteRefer = `[{"table_name":"shared","column_name":["id","label"],"column_type":["int","string"],"row_data":[[7,"good"]]}]`

func startSQLitePoolFixture(t *testing.T, serve http.HandlerFunc) *refertable.ReferTable {
	t.Helper()
	server := httptest.NewServer(serve)
	t.Cleanup(server.Close)
	table, err := refertable.NewReferTable(refertable.RefTbCfg{
		URL: server.URL, Interval: time.Second, UseSQLite: true, SQLiteMemMode: true,
	})
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
			t.Error("worker failed to stop")
		}
	})
	return table
}

func TestSQLiteReferInitializationRequiresSuccessfulUpdate(t *testing.T) {
	var valid atomic.Bool
	table := startSQLitePoolFixture(t, func(w http.ResponseWriter, r *http.Request) {
		if valid.Load() {
			fmt.Fprint(w, validSQLiteRefer)
		} else {
			// Valid JSON, but invalid table schema: download succeeds, update fails.
			fmt.Fprint(w, `[{"table_name":"shared","column_name":["id"],"column_type":[],"row_data":[]}]`)
		}
	})
	if table.InitFinished(150 * time.Millisecond) {
		t.Fatal("failed table update incorrectly marked initialization successful")
	}
	valid.Store(true)
	if !table.InitFinished(3 * time.Second) {
		t.Fatal("valid update did not recover initialization")
	}
	row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
	if !ok || row["label"] != "good" {
		t.Fatalf("recovered query: %v %v", row, ok)
	}
}

func TestSQLiteReferMemoryConcurrentQueries(t *testing.T) {
	table := startSQLitePoolFixture(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, validSQLiteRefer) })
	if !table.InitFinished(2 * time.Second) {
		t.Fatal("initialization failed")
	}
	var failures atomic.Int32
	var workers sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for j := 0; j < 50; j++ {
				row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
				if !ok || row["label"] != "good" {
					failures.Add(1)
					return
				}
				stats := table.Tables().Stats()
				if stats == nil || len(stats.Row) != 1 || stats.Row[0] != 1 {
					failures.Add(1)
					return
				}
			}
		}()
	}
	close(start)
	workers.Wait()
	if n := failures.Load(); n != 0 {
		t.Fatalf("%d concurrent readers lost the initialized in-memory database", n)
	}
}
