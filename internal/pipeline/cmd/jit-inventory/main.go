// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// jit-inventory inventories upstream source, not compatibility test results.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
type function struct {
	Name string `json:"name"`
	location
	Status string `json:"status"`
}
type testEntry struct {
	Name string `json:"name"`
	location
	CaseLabels []string `json:"literal_case_labels,omitempty"`
	Status     string   `json:"status"`
}
type inventory struct {
	Schema           int                `json:"schema"`
	Scope            string             `json:"scope"`
	Functions        []function         `json:"functions"`
	Tests            []testEntry        `json:"tests"`
	DataKitOverrides []registryOverride `json:"datakit_registry_observations,omitempty"`
}

func collect(root string) (inventory, error) {
	out := inventory{Schema: 1, Scope: "pipeline-go source inventory only; labels are candidates, not expanded or executed cases; no compatibility inferred"}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		isRegistry := rel == "ptinput/funcs/all.go"
		if !isRegistry && !strings.HasSuffix(rel, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		loc := func(n ast.Node) location { return location{rel, fset.Position(n.Pos()).Line} }
		if isRegistry {
			ast.Inspect(file, func(n ast.Node) bool {
				v, ok := n.(*ast.ValueSpec)
				if !ok || len(v.Names) != 1 || v.Names[0].Name != "FuncsMap" || len(v.Values) != 1 {
					return true
				}
				m, ok := v.Values[0].(*ast.CompositeLit)
				if !ok {
					return false
				}
				for _, item := range m.Elts {
					kv, ok := item.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if name, ok := literal(kv.Key); ok {
						out.Functions = append(out.Functions, function{name, loc(kv), "unmapped"})
					}
				}
				return false
			})
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isTest(fn) {
				continue
			}
			test := testEntry{Name: fn.Name.Name, location: loc(fn), Status: "unmapped"}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				// Include table labels even when the script is built dynamically.
				if kv, ok := n.(*ast.KeyValueExpr); ok {
					if key, ok := kv.Key.(*ast.Ident); ok && (key.Name == "name" || key.Name == "Name") {
						if label, ok := literal(kv.Value); ok {
							test.CaseLabels = append(test.CaseLabels, label)
						}
					}
				}
				if call, ok := n.(*ast.CallExpr); ok && len(call.Args) > 0 {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Run" {
						if label, ok := literal(call.Args[0]); ok {
							test.CaseLabels = append(test.CaseLabels, label)
						}
					}
				}
				return true
			})
			out.Tests = append(out.Tests, test)
		}
		return nil
	})
	if err != nil {
		return out, err
	}
	if len(out.Functions) == 0 {
		return out, fmt.Errorf("no ptinput/funcs/all.go FuncsMap entries in %s", root)
	}
	sort.Slice(out.Functions, func(i, j int) bool { return out.Functions[i].Name < out.Functions[j].Name })
	return out, nil
}

func literal(expr ast.Expr) (string, bool) {
	v, ok := expr.(*ast.BasicLit)
	if !ok || v.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(v.Value)
	return s, err == nil
}

func isTest(fn *ast.FuncDecl) bool {
	if fn.Recv != nil || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Type.Params == nil {
		return false
	}
	params := fn.Type.Params.List
	if len(params) != 1 {
		return false
	}
	ptr, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := ptr.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "T"
}

func main() {
	root := flag.String("upstream", "", "pipeline-go module directory (required)")
	output := flag.String("out", "", "generated JSON path; defaults to stdout")
	datakit := flag.String("datakit", "", "optional DataKit repository root for registry observations")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "-upstream is required")
		os.Exit(2)
	}
	report, err := collect(*root)
	if err == nil && *datakit != "" {
		report.DataKitOverrides, err = collectOverrides(*datakit)
	}
	if err == nil {
		var data []byte
		data, err = json.MarshalIndent(report, "", "  ")
		if err == nil {
			data = append(data, '\n')
			if *output == "" {
				_, err = os.Stdout.Write(data)
			} else {
				err = os.WriteFile(*output, data, 0o644)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
