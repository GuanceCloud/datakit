// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func newTestScriptManager() *ScriptManager {
	return NewScriptManager(nil, nil)
}

func TestManagerAcquireContextCancellationAndRecovery(t *testing.T) {
	m := newTestScriptManager()
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			m.rw.Lock()
			locked := true
			defer func() {
				if locked {
					m.rw.Unlock()
				}
			}()
			var ctx context.Context
			var cancel context.CancelFunc
			want := context.Canceled
			if mode == "deadline" {
				ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
				want = context.DeadlineExceeded
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			done := make(chan error, 1)
			go func() {
				lease, err := m.AcquireContext(ctx)
				if lease != nil {
					lease.Release()
					done <- errors.New("canceled acquisition returned lease")
					return
				}
				done <- err
			}()
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatalf("got %v want %v", err, want)
				}
			case <-time.After(time.Second):
				t.Fatal("manager wait ignored cancellation")
			}
			m.rw.Unlock()
			locked = false
			lease, err := m.AcquireContext(context.Background())
			if err != nil || lease == nil {
				t.Fatalf("recovery: %v", err)
			}
			lease.Release()
			if !m.rw.TryLock() {
				t.Fatal("read lock leaked")
			}
			m.rw.Unlock()
		})
	}
	if _, err := m.AcquireContext(nil); err == nil {
		t.Fatal("nil context accepted")
	}
	var missing *ScriptManager
	if _, err := missing.AcquireContext(context.Background()); err == nil {
		t.Fatal("missing manager accepted")
	}
}

func TestModuleAdmissionPreservesRoutesAndRetiresChangedClosures(t *testing.T) {
	manager := newTestScriptManager()
	checks, invalidations := 0, 0
	admit := true
	manager.SetJITModuleHooks(func(*pljit.ModuleSnapshot) bool { checks++; return admit },
		func(*pljit.ModuleSnapshot) { invalidations++ })
	sources := map[string]string{"main.p": `use("child.p")`, "child.p": `add_key(version, 1)`}
	load := func() {
		t.Helper()
		if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
			t.Fatal(err)
		}
	}
	load()
	if checks != 2 {
		t.Fatalf("initial checks=%d", checks)
	}
	admit = false
	sources["unrelated.p"] = `exit()`
	load()
	if checks != 3 || invalidations != 0 {
		t.Fatalf("unrelated update checks=%d invalidations=%d", checks, invalidations)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing lease")
	}
	parent, _ := lease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	eligible := lease.JITModuleEligible(parent)
	lease.Release()
	if !eligible {
		t.Fatal("unchanged route was reconsidered")
	}
	sources["child.p"] = `add_key(version, 2)`
	load()
	if checks != 5 || invalidations != 2 {
		t.Fatalf("dependency update checks=%d invalidations=%d", checks, invalidations)
	}
	lease, ok = manager.Acquire()
	if !ok {
		t.Fatal("missing new lease")
	}
	parent, _ = lease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	eligible = lease.JITModuleEligible(parent)
	lease.Release()
	if eligible {
		t.Fatal("changed closure reused old admission")
	}
	load()
	if checks != 5 || invalidations != 2 {
		t.Fatal("identical rejected closure was reconsidered")
	}
}

func TestModuleSnapshotSurvivesMetadataOnlyPublication(t *testing.T) {
	manager := newTestScriptManager()
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"main.p": `exit()`}, nil); err != nil {
		t.Fatal(err)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing lease")
	}
	script, ok := lease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	if !ok {
		lease.Release()
		t.Fatal("missing script")
	}
	before, err := lease.JITModuleSnapshot(point.Logging, script)
	lease.Release()
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := manager.PrepareRemoteUpdate(RemoteManagerUpdate{
		ReplaceDefaults: true, Defaults: map[point.Category]string{point.Logging: "main.p"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := prepared.Commit(); err != nil {
		t.Fatal(err)
	}
	next, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing updated lease")
	}
	defer next.Release()
	after, err := next.JITModuleSnapshot(point.Logging, script)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("metadata-only publication replaced the code snapshot")
	}
}

func TestManagerModuleDependencyClosure(t *testing.T) {
	manager := newTestScriptManager()
	sources := map[string]string{
		"main.p":   `use("child.p"); use("shared.p")`,
		"child.p":  `use("shared.p")`,
		"shared.p": `add_key(value, 1)`,
		"after.p":  `create_point("child", {}, {}, after_use="shared.p")`,
	}
	for i := 0; i < 130; i++ {
		sources[fmt.Sprintf("unrelated%d.p", i)] = `exit()`
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
		t.Fatal(err)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing lease")
	}
	defer lease.Release()
	for _, test := range []struct {
		entry string
		names []string
	}{
		{"main.p", []string{"main.p", "child.p", "shared.p"}},
		{"after.p", []string{"after.p", "shared.p"}},
	} {
		script, ok := lease.Manager().QueryScript(point.Logging, test.entry, struct{}{})
		if !ok {
			t.Fatal("missing entry")
		}
		closure, err := moduleSourceClosure(script.Engine(), sources)
		if err != nil {
			t.Fatal(err)
		}
		if len(closure) != len(test.names) {
			t.Fatalf("unexpected closure: %v", closure)
		}
		for _, name := range test.names {
			if closure[name] != sources[name] {
				t.Fatalf("missing dependency %s", name)
			}
		}
		first, err := lease.JITModuleSnapshot(point.Logging, script)
		if err != nil {
			t.Fatalf("unrelated modules incorrectly counted: %v", err)
		}
		second, err := lease.JITModuleSnapshot(point.Logging, script)
		if err != nil || first != second {
			t.Fatal("snapshot rebuilt during batch lookup")
		}
	}
}

func TestManagerModuleSnapshotRejectsMixedGenerations(t *testing.T) {
	manager := newTestScriptManager()
	sources := map[string]string{"main.p": `use("child.p")`, "child.p": `add_key(version, 1)`}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
		t.Fatal(err)
	}
	oldLease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing lease")
	}
	oldScript, ok := oldLease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	if !ok {
		oldLease.Release()
		t.Fatal("missing script")
	}
	_, err := oldLease.JITModuleSnapshot(point.Logging, oldScript)
	oldLease.Release()
	if err != nil {
		t.Fatal(err)
	}
	// Parent source is unchanged: pointer ownership must still prevent mixing
	// its old dependency bindings with the new generation's dependency snapshot.
	sources["child.p"] = `add_key(version, 2)`
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
		t.Fatal(err)
	}
	next, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing new lease")
	}
	defer next.Release()
	if _, err := next.JITModuleSnapshot(point.Logging, oldScript); err == nil {
		t.Fatal("accepted old parent in new generation")
	}
	current, ok := next.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	if !ok {
		t.Fatal("missing new parent")
	}
	if _, err := next.JITModuleSnapshot(point.Logging, current); err != nil {
		t.Fatal(err)
	}
	if _, err := next.JITModuleSnapshot(point.Metric, current); err == nil {
		t.Fatal("accepted wrong category")
	}
}

func TestUnchangedSourceKeepsRouteAcrossUnrelatedUpdates(t *testing.T) {
	for _, initial := range []bool{false, true} {
		manager := newTestScriptManager()
		const original = "drop_key(original)\n"
		decision := initial
		checks := 0
		manager.SetJITCheck(func(source string) bool {
			if source == original {
				checks++
			}
			return decision
		})
		if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"original.p": original}, nil); err != nil {
			t.Fatal(err)
		}
		decision = !initial
		if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSRemote, map[string]string{"other.p": "drop_key(other)\n"}, nil); err != nil {
			t.Fatal(err)
		}
		lease, ok := manager.Acquire()
		if !ok {
			t.Fatal("missing manager")
		}
		got := lease.JITEligible(original)
		lease.Release()
		if got != initial || checks != 1 {
			t.Fatalf("initial=%v route=%v checks=%d: unchanged source was rerouted", initial, got, checks)
		}
	}
}

func TestScriptManagerCompileFailureRetainsCurrent(t *testing.T) {
	manager := newTestScriptManager()
	invalidated := make(chan string, 1)
	manager.SetJITInvalidate(func(source string) { invalidated <- source })
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"test.p": "drop_key(foo)\n"}, nil); err != nil {
		t.Fatalf("load valid script: %v", err)
	}
	before, ok := manager.QueryScript(point.Logging, "test.p")
	if !ok {
		t.Fatal("query valid script")
	}

	recorded := plstats.NewRecStats("test", "script_manager", nil, 16)
	plstats.SetStats(recorded)
	defer plstats.SetStats(nil)
	err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"test.p": "if"}, nil)
	if err == nil {
		t.Fatal("invalid candidate unexpectedly compiled")
	}
	after, ok := manager.QueryScript(point.Logging, "test.p")
	if !ok || after != before || after.Content() != "drop_key(foo)\n" {
		t.Fatalf("compile failure replaced current script: %#v", after)
	}
	select {
	case source := <-invalidated:
		t.Fatalf("compile failure invalidated active source %q", source)
	default:
	}
	var foundCompileError bool
	for _, event := range recorded.ReadEvents(make([]*plstats.ChangeEvent, 0, 16)) {
		if event.Op == plstats.EventOpCompileError && event.Name == "test.p" && event.CompileError != "" {
			foundCompileError = true
		}
	}
	if !foundCompileError {
		t.Fatal("rejected candidate did not record a compile-error event")
	}
}

func TestScriptManagerBatchLeaseBlocksCommitAndCleanup(t *testing.T) {
	manager := newTestScriptManager()
	invalidated := make(chan string, 1)
	manager.SetJITInvalidate(func(source string) { invalidated <- source })
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"test.p": "drop_key(old)\n"}, nil); err != nil {
		t.Fatalf("load old script: %v", err)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire batch lease")
	}
	old, ok := lease.Manager().QueryScript(point.Logging, "test.p")
	if !ok {
		t.Fatal("query old script")
	}

	updated := make(chan error, 1)
	go func() {
		updated <- manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
			map[string]string{"test.p": "drop_key(new)\n"}, nil)
	}()
	select {
	case err := <-updated:
		t.Fatalf("update crossed active batch lease: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case source := <-invalidated:
		t.Fatalf("source %q invalidated before its batch drained", source)
	default:
	}
	stillOld, ok := lease.Manager().QueryScript(point.Logging, "test.p")
	if !ok || stillOld != old || stillOld.Content() != "drop_key(old)\n" {
		t.Fatal("active batch did not retain its old script")
	}

	lease.Release()
	select {
	case err := <-updated:
		if err != nil {
			t.Fatalf("commit new script: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("update did not commit after batch drained")
	}
	select {
	case source := <-invalidated:
		if source != "drop_key(old)\n" {
			t.Fatalf("invalidated source = %q", source)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("retired script source was not invalidated")
	}
	current, ok := manager.QueryScript(point.Logging, "test.p")
	if !ok || current == old || current.Content() != "drop_key(new)\n" {
		t.Fatalf("new script was not committed: %#v", current)
	}
}

func TestScriptManagerTargetedCommitPreservesUnchangedNamespace(t *testing.T) {
	manager := newTestScriptManager()
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"stable.p": "drop_key(stable)\n"}, nil); err != nil {
		t.Fatalf("load stable script: %v", err)
	}
	stable, ok := manager.QueryScript(point.Logging, "stable.p")
	if !ok {
		t.Fatal("query stable script")
	}

	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSRemote,
		map[string]string{"new.p": "drop_key(new)\n"}, nil); err != nil {
		t.Fatalf("load targeted remote script: %v", err)
	}
	after, ok := manager.QueryScript(point.Logging, "stable.p")
	if !ok || after.Content() != stable.Content() {
		t.Fatal("targeted commit changed an unchanged namespace script")
	}
	if _, ok := manager.QueryScript(point.Logging, "new.p"); !ok {
		t.Fatal("targeted script was not committed")
	}
}

func TestScriptManagerPublishesJITRouteWithCommit(t *testing.T) {
	manager := newTestScriptManager()
	checkStarted := make(chan struct{})
	allowCheck := make(chan struct{})
	manager.SetJITCheck(func(source string) bool {
		if source == "drop_key(new)\n" {
			close(checkStarted)
			<-allowCheck
			return true
		}
		return false
	})

	updated := make(chan error, 1)
	go func() {
		updated <- manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
			map[string]string{"test.p": "drop_key(new)\n"}, nil)
	}()
	select {
	case <-checkStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("candidate JIT check did not start")
	}
	if _, ok := manager.QueryScript(point.Logging, "test.p"); ok {
		t.Fatal("script became visible before its JIT check completed")
	}
	close(allowCheck)
	if err := <-updated; err != nil {
		t.Fatalf("commit checked script: %v", err)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire committed script")
	}
	defer lease.Release()
	script, ok := lease.Manager().QueryScript(point.Logging, "test.p")
	if !ok || !lease.JITEligible(script.Content()) {
		t.Fatal("script and JIT route were not published together")
	}
}

func TestManagerLeaseSnapshotsRelationForWholeBatch(t *testing.T) {
	manager := newTestScriptManager()
	manager.UpdateRelation(1, map[point.Category]map[string]string{
		point.Logging: {"source": "old.p"},
	})
	oldLease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire old relation snapshot")
	}

	updated := make(chan struct{})
	go func() {
		manager.UpdateRelation(2, map[point.Category]map[string]string{
			point.Logging: {"source": "new.p"},
		})
		close(updated)
	}()
	select {
	case <-updated:
		t.Fatal("relation publication crossed an active batch lease")
	case <-time.After(100 * time.Millisecond):
	}
	if name, ok := oldLease.Relation().Query(point.Logging, "source"); !ok || name != "old.p" {
		t.Fatalf("active batch relation changed to %q", name)
	}
	oldLease.Release()
	select {
	case <-updated:
	case <-time.After(5 * time.Second):
		t.Fatal("relation publication did not complete after old batch drained")
	}
	newLease, ok := manager.Acquire()
	if !ok {
		t.Fatal("acquire new relation snapshot")
	}
	defer newLease.Release()
	if name, ok := newLease.Relation().Query(point.Logging, "source"); !ok || name != "new.p" {
		t.Fatalf("new batch relation = %q", name)
	}
}

func TestScriptManagerInvalidatesOnlyAfterLastSourceReference(t *testing.T) {
	manager := newTestScriptManager()
	invalidated := make(chan string, 1)
	manager.SetJITInvalidate(func(source string) { invalidated <- source })
	const source = "drop_key(shared)\n"
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"default.p": source}, nil); err != nil {
		t.Fatalf("load default reference: %v", err)
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSRemote,
		map[string]string{"remote.p": source}, nil); err != nil {
		t.Fatalf("load remote reference: %v", err)
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSRemote, nil, nil); err != nil {
		t.Fatalf("remove remote reference: %v", err)
	}
	select {
	case got := <-invalidated:
		t.Fatalf("source invalidated while still referenced: %q", got)
	default:
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, nil, nil); err != nil {
		t.Fatalf("remove final reference: %v", err)
	}
	select {
	case got := <-invalidated:
		if got != source {
			t.Fatalf("invalidated source = %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("source was not invalidated after its final reference was removed")
	}
}
