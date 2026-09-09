// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput"
)

func TestRemoteUpdatePreservesLocalScriptState(t *testing.T) {
	m := newTestScriptManager()
	defer func() { cleanupManagerState(m.state) }()
	source := `n=cache_get("count"); cache_set("count","retained"); add_key(count,n)`
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"counter.p": source}, nil); err != nil {
		t.Fatal(err)
	}
	before, _ := m.QueryScript(point.Logging, "counter.p")
	run := func() any {
		t.Helper()
		s, ok := m.QueryScript(point.Logging, "counter.p")
		if !ok {
			t.Fatal("missing counter")
		}
		p := ptinput.NewPlPt(point.Logging, "test", nil, nil, time.Now())
		if err := s.Run(p, nil, nil); err != nil {
			t.Fatal(err)
		}
		v, _, _ := p.Get("count")
		return v
	}
	run()
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSRemote, map[string]string{"other.p": `add_key(other,true)`}, nil); err != nil {
		t.Fatal(err)
	}
	after, _ := m.QueryScript(point.Logging, "counter.p")
	if got := run(); got != "retained" || after != before {
		t.Fatalf("remote update replaced local script or lost cache: %v", got)
	}
	if err := m.LoadScriptWithCatChecked(point.Metric, constants.NSDefault, map[string]string{"metric.p": `add_key(value,1)`}, nil); err != nil {
		t.Fatal(err)
	}
	after, _ = m.QueryScript(point.Logging, "counter.p")
	if got := run(); got != "retained" || after != before {
		t.Fatalf("other category update replaced local script or lost cache: %v", got)
	}
	// Explicitly reloading the same namespace/category must retain the legacy
	// reset behavior, even when the script source is unchanged.
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"counter.p": source}, nil); err != nil {
		t.Fatal(err)
	}
	after, _ = m.QueryScript(point.Logging, "counter.p")
	if got := run(); got != nil || after == before {
		t.Fatalf("explicit identical reload did not replace script and reset cache: %v", got)
	}
}

func TestWorkspaceInitialLoadKeepsValidScripts(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{"valid.p": `add_key(ok,true)`, "invalid.p": "if"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	m := newTestScriptManager()
	defer func() { cleanupManagerState(m.state) }()
	m.LoadScriptsFromWorkspace(constants.NSDefault, dir, nil)
	if _, ok := m.QueryScript(point.Logging, "valid.p"); !ok {
		t.Fatal("valid script rejected with invalid sibling")
	}
	if _, ok := m.QueryScript(point.Logging, "invalid.p", struct{}{}); ok {
		t.Fatal("invalid script installed")
	}
}

func TestScriptReplacementDependenciesAbortAndPriority(t *testing.T) {
	m := newTestScriptManager()
	defer func() { cleanupManagerState(m.state) }()
	source := map[string]string{"root.p": `use("child.p")`, "child.p": `add_key(value,1)`, "counter.p": `cache_set("key","value")`}
	load := func(s map[string]string) {
		t.Helper()
		if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, s, nil); err != nil {
			t.Fatal(err)
		}
	}
	load(source)
	root, _ := m.QueryScript(point.Logging, "root.p")
	counter, _ := m.QueryScript(point.Logging, "counter.p")
	source["child.p"] = `add_key(value,2)`
	load(source)
	changed, _ := m.QueryScript(point.Logging, "root.p")
	same, _ := m.QueryScript(point.Logging, "counter.p")
	if root == changed || counter == same {
		t.Fatal("namespace reload did not replace every script")
	}
	counter = same
	pt := ptinput.NewPlPt(point.Logging, "test", nil, nil, time.Now())
	if err := changed.Run(pt, nil, nil); err != nil {
		t.Fatal(err)
	}
	if v, _, _ := pt.Get("value"); v != int64(2) {
		t.Fatalf("stale linked child: %v", v)
	}
	update := RemoteManagerUpdate{ReplaceScripts: true, Scripts: map[point.Category]map[string]string{point.Logging: {"counter.p": `add_key(remote,true)`}}}
	prepared, err := m.PrepareRemoteUpdate(update)
	if err != nil {
		t.Fatal(err)
	}
	prepared.Abort()
	prepared.Abort()
	got, _ := m.QueryScript(point.Logging, "counter.p")
	if got != counter {
		t.Fatal("Abort changed published script")
	}
	prepared, err = m.PrepareRemoteUpdate(update)
	if err != nil {
		t.Fatal(err)
	}
	if err = prepared.Commit(); err != nil {
		t.Fatal(err)
	}
	got, _ = m.QueryScript(point.Logging, "counter.p")
	if got.NS() != constants.NSRemote {
		t.Fatal("remote priority lost")
	}
	prepared, err = m.PrepareRemoteUpdate(RemoteManagerUpdate{ReplaceScripts: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = prepared.Commit(); err != nil {
		t.Fatal(err)
	}
	got, _ = m.QueryScript(point.Logging, "counter.p")
	if got != counter {
		t.Fatal("remote removal did not restore original local instance")
	}
	source["child.p"] = "if"
	if err = m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, source, nil); err == nil {
		t.Fatal("invalid hot update accepted")
	}
	got, _ = m.QueryScript(point.Logging, "root.p")
	if got != changed {
		t.Fatal("failed update replaced working graph")
	}
}

func TestActualInitializationWithBadWorkspaceSibling(t *testing.T) {
	if os.Getenv("DK_SCRIPT_INIT_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestActualInitializationWithBadWorkspaceSibling$", "-test.timeout=20s")
		cmd.Env = append(os.Environ(), "DK_SCRIPT_INIT_CHILD=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("initialization subprocess: %v\n%s", err, out)
		}
		return
	}
	dir := t.TempDir()
	pipelineDir := filepath.Join(dir, "pipeline")
	if err := os.Mkdir(pipelineDir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{"valid.p": `add_key(ok,true)`, "invalid.p": "if"} {
		if err := os.WriteFile(filepath.Join(pipelineDir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	prepared := &PreparedPlValServices{}
	defer prepared.Abort()
	if err := InitPlValPrepared(nil, nil, nil, dir, prepared); err != nil {
		t.Fatal(err)
	}
	m, ok := GetManager()
	if !ok {
		t.Fatal("manager not published")
	}
	defer cleanupManagerState(m.state)
	script, ok := m.QueryScript(point.Logging, "valid.p")
	if !ok {
		t.Fatal("valid script missing after actual initialization")
	}
	pt := ptinput.NewPlPt(point.Logging, "test", nil, nil, time.Now())
	if err := script.Run(pt, nil, nil); err != nil {
		t.Fatal(err)
	}
	if v, _, _ := pt.Get("ok"); v != true {
		t.Fatalf("valid script not executed: %v", v)
	}
}

func TestWorkspaceFirstImportPerNamespace(t *testing.T) {
	for _, ns := range []string{constants.NSGitRepo, constants.NSConfd} {
		t.Run(ns, func(t *testing.T) {
			m := newTestScriptManager()
			defer func() { cleanupManagerState(m.state) }()
			if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"builtin.p": `add_key(builtin,true)`}, nil); err != nil {
				t.Fatal(err)
			}
			builtin, _ := m.QueryScript(point.Logging, "builtin.p")
			dir := t.TempDir()
			for name, source := range map[string]string{"valid.p": `add_key(ok,true)`, "invalid.p": "if"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			// Both GitRepo reload and Confd import call this public workspace entry.
			m.LoadScriptsFromWorkspace(ns, dir, nil)
			valid, ok := m.QueryScript(point.Logging, "valid.p", struct{}{})
			if !ok {
				t.Fatal("first namespace import rejected valid sibling")
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, nil, time.Now())
			if err := valid.Run(pt, nil, nil); err != nil {
				t.Fatal(err)
			}
			if v, _, _ := pt.Get("ok"); v != true {
				t.Fatalf("valid script did not execute: %v", v)
			}
			if _, ok := m.QueryScript(point.Logging, "invalid.p", struct{}{}); ok {
				t.Fatal("invalid script published")
			}
			current, _ := m.QueryScript(point.Logging, "builtin.p")
			if current != builtin {
				t.Fatal("import replaced default namespace")
			}
			// The next load is a transaction: bad siblings retain the working generation.
			if err := os.WriteFile(filepath.Join(dir, "valid.p"), []byte(`add_key(ok,false)`), 0600); err != nil {
				t.Fatal(err)
			}
			m.LoadScriptsFromWorkspace(ns, dir, nil)
			current, _ = m.QueryScript(point.Logging, "valid.p", struct{}{})
			if current != valid {
				t.Fatal("failed hot reload replaced working namespace")
			}
			if err := os.Remove(filepath.Join(dir, "invalid.p")); err != nil {
				t.Fatal(err)
			}
			m.LoadScriptsFromWorkspace(ns, dir, nil)
			current, _ = m.QueryScript(point.Logging, "valid.p", struct{}{})
			if current == valid {
				t.Fatal("repaired namespace did not publish")
			}
		})
	}
}
