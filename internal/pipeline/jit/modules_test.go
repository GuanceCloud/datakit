// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import "testing"

func TestScopedModuleSnapshotIdentity(t *testing.T) {
	sources := map[string]string{"main.p": `use("child.p")`, "child.p": `exit()`}
	build := func(ns, cat, entry string, source map[string]string) *ModuleSnapshot {
		t.Helper()
		snapshot, err := NewScopedModuleSnapshot(ns, cat, entry, source)
		if err != nil {
			t.Fatal(err)
		}
		return snapshot
	}
	original := build("remote", "logging", "main.p", sources)
	same := build("remote", "logging", "main.p", map[string]string{"child.p": `exit()`, "main.p": `use("child.p")`})
	if original.key != same.key {
		t.Fatal("same instance identity differs")
	}
	for _, other := range []*ModuleSnapshot{
		build("local", "logging", "main.p", sources),
		build("remote", "metric", "main.p", sources),
		build("remote", "logging", "child.p", sources),
		build("remot", "elogging", "main.p", sources),
	} {
		if original.key == other.key {
			t.Fatal("different owner or entry shares state identity")
		}
	}
	sources["child.p"] = `add_key(version, 2)`
	if original.key == build("remote", "logging", "main.p", sources).key {
		t.Fatal("dependency update not reflected")
	}
	if original.sources["child.p"] != `exit()` {
		t.Fatal("caller mutated snapshot")
	}
}
