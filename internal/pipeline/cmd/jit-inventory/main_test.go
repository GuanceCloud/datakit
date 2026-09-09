// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectPreservesAliasesAndUnmappedCases(t *testing.T) {
	root := t.TempDir()
	write := func(name, source string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("ptinput/funcs/all.go", `package funcs
var FuncsMap = map[string]any{"valid_json": Valid, "vaild_json": Valid}
var FuncsCheckMap = map[string]any{"ignored": Check}
`)
	write("ptinput/funcs/example_test.go", `package funcs
import "testing"
func TestExample(t *testing.T) {
 cases := []struct{name string}{{name: "same"}, {name: "same"}}
 for _, tc := range cases { t.Run(tc.name, func(t *testing.T){}) }
 t.Run("explicit", func(t *testing.T){})
}
func BenchmarkExample(b *testing.B) {}
func TestHelper() {}
`)
	got, err := collect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Functions) != 2 || got.Functions[0].Name != "vaild_json" || got.Functions[1].Name != "valid_json" {
		t.Fatalf("aliases: %+v", got.Functions)
	}
	if len(got.Tests) != 1 || !reflect.DeepEqual(got.Tests[0].CaseLabels, []string{"same", "same", "explicit"}) {
		t.Fatalf("tests: %+v", got.Tests)
	}
	if got.Tests[0].Status != "unmapped" || got.Functions[0].Status != "unmapped" {
		t.Fatal("inventory must not infer coverage")
	}
	for _, f := range got.Functions {
		if f.File != "ptinput/funcs/all.go" || f.Line != 2 {
			t.Fatalf("location: %+v", f)
		}
	}
	again, err := collect(root)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(again)
	if string(a) != string(b) {
		t.Fatal("non-deterministic inventory")
	}
}

func TestCollectRejectsWrongRootAndInvalidSource(t *testing.T) {
	root := t.TempDir()
	if _, err := collect(root); err == nil {
		t.Fatal("empty root accepted")
	}
	if err := os.WriteFile(filepath.Join(root, "broken_test.go"), []byte("package"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := collect(root); err == nil {
		t.Fatal("parse error swallowed")
	}
}
