// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectOverrides(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "pipeline")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := `package pipeline
import upstream "github.com/GuanceCloud/pipeline-go/ptinput/funcs"
func configure() {
 upstream.FuncsMap["http_request"] = disabled
 once.Do(func() { upstream.FuncsCheckMap["grok"] = checker })
 upstream.FuncsMap[key] = dynamic
 delete(upstream.FuncsMap, "removed")
 upstream.SetNetFilter(true, nil, nil)
 upstream.FuncsMap = replacement
 _ = upstream.FuncsMap["read_not_write"]
 other.FuncsMap["unrelated"] = value
}
`
	for _, name := range []string{"config.go", "config_test.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := collectOverrides(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 6 {
		t.Fatalf("observations: %+v", got)
	}
	wantKinds := []string{"entry_assignment", "entry_assignment", "entry_assignment", "entry_delete", "network_policy", "registry_assignment"}
	wantNames := []string{"http_request", "grok", "", "removed", "http_request", ""}
	for i, item := range got {
		if item.Kind != wantKinds[i] || item.Name != wantNames[i] || item.DynamicKey != (i == 2) || item.Status != "unmapped" || item.Enclosing != "configure" || item.File != "internal/pipeline/config.go" || item.Line != i+4 {
			t.Fatalf("observation %d: %+v", i, item)
		}
	}
	again, err := collectOverrides(root)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatalf("unstable inventory: %v", err)
	}
}

func TestCollectOverridesRejectsInvalidInputs(t *testing.T) {
	root := t.TempDir()
	if _, err := collectOverrides(root); err == nil {
		t.Fatal("missing internal directory accepted")
	}
	dir := filepath.Join(root, "internal")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"package", `package p; import . "github.com/GuanceCloud/pipeline-go/ptinput/funcs"`} {
		if err := os.WriteFile(filepath.Join(dir, "invalid.go"), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := collectOverrides(root); err == nil {
			t.Fatalf("accepted %q", source)
		}
	}
}
