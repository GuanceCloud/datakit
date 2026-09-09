// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type snapshotIPDB struct {
	label string
	onGeo func()
}

func (*snapshotIPDB) Init(string, map[string]string) {}
func (db *snapshotIPDB) Geo(string) (*ipdb.IPdbRecord, error) {
	if db.onGeo != nil {
		db.onGeo()
	}
	return &ipdb.IPdbRecord{City: db.label}, nil
}
func (db *snapshotIPDB) GeoWithChecker(ip string, check ipdb.CheckData) (*ipdb.IPdbRecord, error) {
	r, err := db.Geo(ip)
	if check != nil && err == nil {
		r = check(r)
	}
	return r, err
}
func (db *snapshotIPDB) SearchIsp(string) string { return db.label }

func TestRunPlIPDBReplacementDuringBatch(t *testing.T) {
	originalManager, _ := plval.GetManager()
	originalDB, _ := plval.GetIPDB()
	t.Cleanup(func() { plval.SetManager(originalManager); plval.SetIPDB(originalDB) })
	manager := plval.NewScriptManager(nil, nil)
	// This tests the Go production adapter. Native host generation tests are
	// separate; do not count compile-time routing as a native parity result.
	manager.SetJITCheck(func(string) bool { return false })
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"snapshot.p": `geoip(ip); add_key(after,true)`}, nil); err != nil {
		t.Fatal(err)
	}
	manager.UpdateDefaultScript(map[point.Category]string{point.Logging: "snapshot.p"})
	plval.SetManager(manager)
	entered, release := make(chan struct{}), make(chan struct{})
	var once, releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	plval.SetIPDB(&snapshotIPDB{label: "old", onGeo: func() {
		once.Do(func() { close(entered); <-release })
	}})
	newPoints := func() []*point.Point {
		pts := make([]*point.Point, 10)
		for i := range pts {
			pts[i] = point.NewPoint("snapshot", point.NewKVs(map[string]any{"ip": "8.8.8.8", "sequence": int64(i)}))
		}
		return pts
	}
	type result struct {
		value *ScriptResult
		err   error
	}
	done := make(chan result, 1)
	go func() { v, err := RunPl(point.Logging, newPoints(), nil); done <- result{v, err} }()
	t.Cleanup(func() {
		unblock()
		// Even on assertion failure, don't restore globals before the old run exits.
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("old batch did not exit")
		}
	})
	select {
	case <-entered:
	case r := <-done:
		done <- r
		t.Fatalf("old batch never entered geoip: %v", r.err)
	case <-time.After(3 * time.Second):
		t.Fatal("old batch never entered geoip")
	}
	jitInitMu.Lock()
	plval.SetIPDB(&snapshotIPDB{label: "new"})
	jitInitMu.Unlock()
	check := func(r result, want string) {
		t.Helper()
		if r.err != nil || r.value == nil {
			t.Fatalf("batch failed: %v", r.err)
		}
		if len(r.value.Pts()) != 10 || len(r.value.PtsOffload()) != 0 || len(r.value.PtsCreated()) != 0 {
			t.Fatalf("unexpected batch output: %+v", r.value)
		}
		for i, pt := range r.value.Pts() {
			if pt.Get("city") != want || pt.Get("after") != true || pt.Get("sequence") != int64(i) {
				t.Fatalf("record %d expected %s: %v", i, want, pt)
			}
		}
	}
	// Execute the new batch while the old batch is still blocked in its first call.
	newResult, err := RunPl(point.Logging, newPoints(), nil)
	check(result{newResult, err}, "new")
	unblock()
	select {
	case old := <-done:
		done <- old // retain completion for cleanup
		check(old, "old")
	case <-time.After(3 * time.Second):
		t.Fatal("old batch did not finish")
	}
}

func TestIPDBSnapshotScriptExecution(t *testing.T) {
	originalManager, _ := plval.GetManager()
	originalDB, _ := plval.GetIPDB()
	t.Cleanup(func() { plval.SetManager(originalManager); plval.SetIPDB(originalDB) })
	plval.SetManager(plval.NewScriptManager(nil, nil))
	capture := func(db ipdb.IPdb) ipdb.IPdb {
		t.Helper()
		plval.SetIPDB(db)
		manager, runner, offload, _, snapshot, err := acquirePipelineLeasesContext(context.Background())
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
	empty := capture(nil)
	old := capture(&snapshotIPDB{label: "old"})
	next := capture(&snapshotIPDB{label: "new"})
	script, err := NewPlScriptSimple(point.Logging, "ipdb-snapshot.p", `geoip(ip); add_key(after,true)`)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		snapshot ipdb.IPdb
		want     any
	}{
		{"old_batch", old, "old"}, {"new_batch", next, "new"}, {"absent_at_start", empty, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				pt := point.NewPoint("snapshot", point.NewKVs(map[string]any{"ip": "8.8.8.8", "sequence": int64(i)}))
				run := pointRun{point: pt, script: script, ipdbCaptured: true, ipdbSnapshot: test.snapshot}
				runPipelineGoContext(context.Background(), point.Logging, &run, nil)
				if run.output == nil || run.dropped || len(run.created) != 0 {
					t.Fatalf("unexpected result: %+v", run)
				}
				if got := run.output.Get("city"); got != test.want {
					t.Fatalf("city=%v want=%v", got, test.want)
				}
				if run.output.Get("after") != true || run.output.Get("sequence") != int64(i) {
					t.Fatal("lost input or subsequent statement")
				}
			}
		})
	}
}
