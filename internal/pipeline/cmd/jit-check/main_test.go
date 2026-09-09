// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectUsesLinkedDependencyClosure(t *testing.T) {
	dir := t.TempDir()
	write := func(name, source string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("main.p", "use(\"child.p\")")
	write("child.p", "add_key(a, 1)")
	write("unrelated.p", "a=1")
	first, err := inspect(dir, "default", "logging", "main.p", "", "")
	if err != nil {
		t.Fatal(err)
	}
	write("unrelated.p", "a=2")
	second, err := inspect(dir, "default", "logging", "main.p", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Allow != second.Allow {
		t.Fatal("unrelated script invalidated approval")
	}
	write("child.p", "add_key(a, 2)")
	third, err := inspect(dir, "default", "logging", "main.p", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Allow.BundleSHA256 == third.Allow.BundleSHA256 {
		t.Fatal("dependency change did not invalidate approval")
	}
}
