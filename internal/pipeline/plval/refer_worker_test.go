// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

func TestReferWorkerBindsCreatingGeneration(t *testing.T) {
	oldGlobal, _ := GetRefTb()
	defer SetRefTb(oldGlobal)
	var oldCalls, newCalls atomic.Int32
	server := func(calls *atomic.Int32) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); fmt.Fprint(w, `[]`) }))
	}
	oldServer, newServer := server(&oldCalls), server(&newCalls)
	defer oldServer.Close()
	defer newServer.Close()
	oldTable, err := refertable.NewReferTable(refertable.RefTbCfg{URL: oldServer.URL, Interval: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	newTable, err := refertable.NewReferTable(refertable.RefTbCfg{URL: newServer.URL, Interval: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	SetRefTb(oldTable)
	worker := referPullTask(oldTable)
	// Publish the next table before the scheduled old worker starts.
	SetRefTb(newTable)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- worker(ctx) }()
	if !oldTable.InitFinished(2 * time.Second) {
		t.Fatal("old worker did not initialize its own table")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("old worker did not stop")
	}
	if oldCalls.Load() < 1 || newCalls.Load() != 0 {
		t.Fatalf("wrong generation requests old=%d new=%d", oldCalls.Load(), newCalls.Load())
	}
	if current, ok := GetRefTb(); !ok || current != newTable {
		t.Fatal("old worker changed publication")
	}
	if err := referPullTask(nil)(context.Background()); err == nil {
		t.Fatal("nil worker accepted")
	}
}

func TestReferWorkerReplacementStopsAndDrainsOldGeneration(t *testing.T) {
	oldGlobal, _ := GetRefTb()
	replaceReferWorker(nil)
	t.Cleanup(func() {
		replaceReferWorker(nil)
		SetRefTb(oldGlobal)
	})

	var oldCalls, newCalls atomic.Int32
	server := func(calls *atomic.Int32) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			fmt.Fprint(w, `[]`)
		}))
	}
	oldServer, newServer := server(&oldCalls), server(&newCalls)
	defer oldServer.Close()
	defer newServer.Close()

	const interval = 20 * time.Millisecond
	oldTable, err := refertable.NewReferTable(refertable.RefTbCfg{URL: oldServer.URL, Interval: interval})
	if err != nil {
		t.Fatal(err)
	}
	SetRefTb(oldTable)
	replaceReferWorker(oldTable)
	if !oldTable.InitFinished(2 * time.Second) {
		t.Fatal("old generation did not initialize")
	}

	newTable, err := refertable.NewReferTable(refertable.RefTbCfg{URL: newServer.URL, Interval: interval})
	if err != nil {
		t.Fatal(err)
	}
	SetRefTb(newTable)
	replaceReferWorker(newTable)
	oldAfterDrain := oldCalls.Load()
	if !newTable.InitFinished(2 * time.Second) {
		t.Fatal("new generation did not initialize")
	}
	time.Sleep(4 * interval)
	if got := oldCalls.Load(); got != oldAfterDrain {
		t.Fatalf("retired refer worker kept pulling: before=%d after=%d", oldAfterDrain, got)
	}
	if newCalls.Load() < 1 {
		t.Fatal("published refer worker never pulled")
	}

	replaceReferWorker(nil)
	newAfterDrain := newCalls.Load()
	time.Sleep(4 * interval)
	if got := newCalls.Load(); got != newAfterDrain {
		t.Fatalf("cleared refer worker kept pulling: before=%d after=%d", newAfterDrain, got)
	}
}

func TestReferWorkerReplacementClosesStoreAfterFinalLease(t *testing.T) {
	replaceReferWorker(nil)
	t.Cleanup(func() { replaceReferWorker(nil) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `[{
  "table_name":"shared",
  "column_name":["id","value"],
  "column_type":["int","string"],
  "row_data":[[7,"ready"]]
}]`)
	}))
	defer server.Close()
	newTable := func() *refertable.ReferTable {
		t.Helper()
		table, err := refertable.NewReferTable(refertable.RefTbCfg{
			URL: server.URL, Interval: time.Second, UseSQLite: true, SQLiteMemMode: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		return table
	}

	old := newTable()
	replaceReferWorker(old)
	if !old.InitFinished(3 * time.Second) {
		t.Fatal("old SQLite generation did not initialize")
	}
	lease, ok := AcquireRefTb()
	if !ok || lease.Tables() != old.Tables() {
		t.Fatal("failed to pin old refer generation")
	}

	next := newTable()
	replaceReferWorker(next)
	if old.Tables().Stats() == nil {
		t.Fatal("old backing store closed while a batch lease was active")
	}
	if row, found := lease.Tables().Query("shared", []string{"id"}, []any{int64(7)}, nil); !found || row["value"] != "ready" {
		t.Fatalf("leased old generation stopped serving queries: %#v/%v", row, found)
	}
	lease.Release()

	deadline := time.Now().Add(3 * time.Second)
	for old.Tables().Stats() != nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if old.Tables().Stats() != nil {
		t.Fatal("retired SQLite generation remained open after its final lease")
	}
	if !next.InitFinished(3*time.Second) || next.Tables().Stats() == nil {
		t.Fatal("new generation was not independently usable")
	}
}

func TestPreparedPlValServicesAbortClosesReferWithoutPublishing(t *testing.T) {
	replaceReferWorker(nil)
	t.Cleanup(func() { replaceReferWorker(nil) })

	table, err := refertable.NewReferTable(refertable.RefTbCfg{
		URL:           "http://127.0.0.1:1/unused",
		Interval:      time.Hour,
		UseSQLite:     true,
		SQLiteMemMode: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if table.Tables().Stats() == nil {
		t.Fatal("candidate refer store was not open")
	}
	prepared := &PreparedPlValServices{refer: table}
	prepared.Abort()
	prepared.Abort()

	if table.Tables().Stats() != nil {
		t.Fatal("aborted candidate refer store remained open")
	}
	if _, ok := GetRefTb(); ok {
		t.Fatal("aborted candidate refer table was published")
	}
	if _, ok := AcquireRefTb(); ok {
		t.Fatal("aborted candidate refer worker generation was published")
	}
}
