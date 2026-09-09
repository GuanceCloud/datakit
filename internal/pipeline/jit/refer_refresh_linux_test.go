// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

func TestReferTableSharedRefreshWithoutRecompile(t *testing.T) {
	runReferRefreshLifecycle(t, false)
}

func TestReferTableRefreshFailuresRetainLastGood(t *testing.T) {
	runReferRefreshLifecycle(t, true)
}

func TestReferTableSQLiteRefreshLifecycle(t *testing.T) {
	runSQLiteReferSubprocess(t, false)
}

func TestReferTableSQLiteDiskRefreshLifecycle(t *testing.T) {
	runSQLiteReferSubprocess(t, true)
}

func TestReferTableSQLiteDiskUpstreamOnly(t *testing.T) {
	runSQLiteReferSubprocess(t, true, true)
}

func TestReferTableSQLiteReadTransactionRecovery(t *testing.T) {
	runSQLiteReferSubprocess(t, true, true, true)
}

func runSQLiteReferSubprocess(t *testing.T, disk bool, upstream ...bool) {
	t.Helper()
	// The pinned upstream ReferTable exposes no Close for its private sql.DB.
	// Isolate this backend so process exit reclaims its connection pool.
	const childKey = "DATAKIT_JIT_SQLITE_REFER_CHILD"
	if os.Getenv(childKey) == "1" {
		runReferRefreshLifecycle(t, true, true, disk, len(upstream) > 0 && upstream[0], len(upstream) > 1 && upstream[1])
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1")
	cmd.Env = append(os.Environ(), childKey+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("SQLite lifecycle: %v\n%s", err, output)
	}
}

func runReferRefreshLifecycle(t *testing.T, failures bool, sqlite ...bool) {
	t.Helper()
	upstreamOnly := len(sqlite) > 2 && sqlite[2]
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" && !upstreamOnly {
		t.Fatal("requires real runtime")
	}
	var version atomic.Int64
	version.Store(1)
	var mode atomic.Int64
	requests := make(chan int64, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentMode := mode.Load()
		select {
		case requests <- currentMode:
		default:
		}
		if currentMode == 1 {
			http.Error(w, "temporary outage", http.StatusServiceUnavailable)
			return
		}
		if currentMode == 2 {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `[{broken`)
			return
		}
		if currentMode == 3 {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `[]`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"table_name":"shared","column_name":["id","label"],"column_type":["int","string"],"row_data":[[7,"v%d"]]}]`, version.Load())
	}))
	defer server.Close()
	useSQLite := len(sqlite) > 0 && sqlite[0]
	disk := len(sqlite) > 1 && sqlite[1]
	dbPath := ""
	if disk {
		dbPath = filepath.Join(t.TempDir(), "refer.sqlite")
	}
	table, err := refertable.NewReferTable(refertable.RefTbCfg{URL: server.URL, Interval: time.Second, UseSQLite: useSQLite, SQLiteMemMode: useSQLite && !disk, DBPath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); table.PullWorker(ctx) }()
	defer func() {
		cancel()
		select {
		case <-workerDone:
			if closer, ok := any(table).(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					t.Errorf("close refer store: %v", err)
				}
				if useSQLite {
					if _, found := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil); found {
						t.Error("closed SQLite still served a row")
					}
					if table.Tables().Stats() != nil {
						t.Error("closed SQLite still served stats")
					}
				}
				if err := closer.Close(); err != nil {
					t.Errorf("repeat close refer store: %v", err)
				}
			}
		case <-time.After(5 * time.Second):
			t.Error("refer worker did not stop")
		}
	}()
	if !table.InitFinished(5 * time.Second) {
		t.Fatal("refer initialization timed out")
	}
	var runner *Runner
	if !upstreamOnly {
		runner, err = NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, func() (*HostCompat, error) { return NewPipelineGoHostWithReferObserved(nil, table.Tables(), nil), nil })
		if err != nil {
			t.Fatal(err)
		}
	}
	defer runner.Close()
	scripts := []string{`query_refer_table("shared","id",7); add_key(script,"one")`, `mquery_refer_table("shared",["id"],[7]); add_key(script,"two")`}
	if upstreamOnly {
		scripts = nil
	}
	for _, source := range scripts {
		if check := runner.Check(source); check.Route != RouteJITWithHost {
			t.Fatalf("expected shared host route: %+v", check)
		}
	}
	run := func(want string) {
		t.Helper()
		if upstreamOnly {
			row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
			if !ok || row["label"] != want {
				t.Fatalf("upstream-only row=%v want%s", row, want)
			}
			return
		}
		for _, source := range scripts {
			got := runHostCompatPoint(t, runner, source, Point{Version: 1, Category: "logging", Measurement: "refer", Fields: map[string]any{"message": "test"}})
			if got.Fields["label"] != want {
				t.Fatalf("shared table output=%+v want%s", got.Fields, want)
			}
		}
	}
	run("v1")
	if len(sqlite) > 3 && sqlite[3] {
		// A separate read transaction deterministically blocks a writer commit.
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		var label string
		if err := tx.QueryRow("SELECT label FROM shared WHERE id=7").Scan(&label); err != nil || label != "v1" {
			t.Fatalf("read lock setup: %s %v", label, err)
		}
		version.Store(2)
		mode.Store(4)
		seen := 0
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for seen < 2 {
			select {
			case request := <-requests:
				if request == 4 {
					seen++
				}
			case <-deadline.C:
				t.Fatal("writer did not attempt refresh while reader held")
			}
		}
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		// No native runtime has been loaded in this reproduction.
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		recovery := time.NewTimer(5 * time.Second)
		defer recovery.Stop()
		for {
			row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
			if ok && row["label"] == "v2" {
				return
			}
			select {
			case <-tick.C:
			case <-recovery.C:
				t.Fatal("upstream SQLite failed to recover after reader released")
			}
		}
	}
	if failures {
		for _, failure := range []int64{1, 2} {
			mode.Store(failure)
			// PullWorker is sequential. The second request proves the first
			// response has finished processing, not just reached the server.
			seen := 0
			deadline := time.NewTimer(5 * time.Second)
			for seen < 2 {
				select {
				case request := <-requests:
					if request == failure {
						seen++
					}
				case <-deadline.C:
					t.Fatal("failure refresh was not processed")
				}
			}
			deadline.Stop()
			run("v1")
		}
		mode.Store(3)
		deadline := time.NewTimer(5 * time.Second)
		tick := time.NewTicker(10 * time.Millisecond)
		defer deadline.Stop()
		defer tick.Stop()
		for {
			if _, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil); !ok {
				break
			}
			select {
			case <-tick.C:
			case <-deadline.C:
				t.Fatal("empty refresh did not delete old table")
			}
		}
		for _, source := range scripts {
			for _, existing := range []bool{false, true} {
				fields := map[string]any{"message": "test"}
				if existing {
					fields["label"] = "keep-existing"
				}
				got := runHostCompatPoint(t, runner, source, Point{Version: 1, Category: "logging", Measurement: "refer", Fields: fields})
				value, present := got.Fields["label"]
				if present != existing || (existing && value != "keep-existing") {
					t.Fatalf("no-match changed fields: %+v", got.Fields)
				}
				if got.Fields["script"] == nil {
					t.Fatal("missing table stopped later statements")
				}
			}
		}
	}
	version.Store(2)
	mode.Store(0)
	deadline := time.After(5 * time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
		if ok && row["label"] == "v2" {
			break
		}
		select {
		case <-tick.C:
		case <-deadline:
			t.Fatal("shared table refresh timed out")
		}
	}
	// Same sources and runner, no invalidation or explicit recompilation.
	run("v2")
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
	row, ok := table.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil)
	if !ok || row["label"] != "v2" {
		t.Fatal("runner close damaged externally owned refer table")
	}
}
