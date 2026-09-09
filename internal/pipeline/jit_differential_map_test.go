// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type differentialMap struct {
	Mappings []struct {
		Functions         []string `json:"functions"`
		DifferentialFile  string   `json:"differential_file"`
		DifferentialTest  string   `json:"differential_test"`
		DifferentialTests []string `json:"differential_tests"`
		Verification      struct {
			Result string `json:"result"`
		} `json:"verification"`
	} `json:"mappings"`
}

type pipelineGoInventory struct {
	Functions []struct {
		Name string `json:"name"`
	} `json:"functions"`
}

var differentialTestName = regexp.MustCompile(`\bTest[A-Za-z0-9_]+\b`)

// The reviewed map is a release artifact, not prose only. Keep it connected to
// real tests so renamed/deleted differential coverage cannot leave a false
// 76/76 claim behind.
func TestJITDifferentialMapReferencesExistingTests(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "pipeline-go-differential-map.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report differentialMap
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Mappings) == 0 {
		t.Fatal("differential map has no mappings")
	}
	for index, mapping := range report.Mappings {
		if len(mapping.Functions) == 0 || mapping.DifferentialFile == "" || (mapping.DifferentialTest == "" && len(mapping.DifferentialTests) == 0) {
			t.Fatalf("mapping %d lacks functions, file or test: %+v", index, mapping)
		}
		if !strings.Contains(mapping.Verification.Result, "passed") {
			t.Fatalf("mapping %d %v has no current passing verification: %q", index, mapping.Functions, mapping.Verification.Result)
		}
		declared := map[string]bool{}
		localFiles := 0
		for _, name := range strings.FieldsFunc(mapping.DifferentialFile, func(r rune) bool { return r == ';' || r == ',' }) {
			name = strings.TrimSpace(name)
			if !strings.HasPrefix(name, "internal/") || !strings.HasSuffix(name, ".go") {
				continue
			}
			localFiles++
			path := filepath.Join("..", "..", filepath.FromSlash(name))
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("mapping %d %v references unreadable %s: %v", index, mapping.Functions, name, err)
			}
			for _, declaration := range file.Decls {
				if function, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(function.Name.Name, "Test") {
					declared[function.Name.Name] = true
				}
			}
		}
		if localFiles == 0 {
			continue // Rust/AOT mappings are verified in that repository.
		}
		references := append([]string{mapping.DifferentialTest}, mapping.DifferentialTests...)
		for _, name := range differentialTestName.FindAllString(strings.Join(references, ";"), -1) {
			if !declared[name] {
				t.Errorf("mapping %d %v references missing %s in %q", index, mapping.Functions, name, mapping.DifferentialFile)
			}
		}
	}
}

// Keep the reviewed compatibility evidence tied to the generated upstream
// denominator. Merely checking references inside the reviewed map would stay
// green if pipeline-go added a builtin that nobody mapped.
func TestJITDifferentialMapCoversUpstreamBuiltinInventory(t *testing.T) {
	read := func(name string, target any) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
	}
	var inventory pipelineGoInventory
	var report differentialMap
	read("pipeline-go-inventory.json", &inventory)
	read("pipeline-go-differential-map.json", &report)

	upstream := make(map[string]struct{}, len(inventory.Functions))
	for _, function := range inventory.Functions {
		if function.Name == "" {
			t.Fatal("upstream inventory contains an empty builtin name")
		}
		if _, exists := upstream[function.Name]; exists {
			t.Fatalf("upstream inventory contains duplicate builtin %q", function.Name)
		}
		upstream[function.Name] = struct{}{}
	}
	mapped := make(map[string]struct{}, len(upstream))
	for _, mapping := range report.Mappings {
		for _, function := range mapping.Functions {
			mapped[function] = struct{}{}
		}
	}
	for function := range upstream {
		if _, exists := mapped[function]; !exists {
			t.Errorf("upstream builtin %q has no reviewed differential mapping", function)
		}
	}
	for function := range mapped {
		if _, exists := upstream[function]; !exists {
			t.Errorf("reviewed differential mapping names non-upstream builtin %q", function)
		}
	}
	if len(upstream) != len(mapped) {
		t.Fatalf("reviewed builtin coverage=%d/%d", len(mapped), len(upstream))
	}
}
