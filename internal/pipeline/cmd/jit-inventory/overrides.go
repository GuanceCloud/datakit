// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Observations identify source sites, not active configuration or compatibility.
// No alias/dataflow analysis is attempted; build constraints remain unfiltered.
type registryOverride struct {
	location
	Enclosing  string `json:"enclosing_function"`
	Kind       string `json:"kind"`
	Registry   string `json:"registry,omitempty"`
	Name       string `json:"name,omitempty"`
	DynamicKey bool   `json:"dynamic_key,omitempty"`
	Status     string `json:"status"`
}

func collectOverrides(root string) ([]registryOverride, error) {
	base := filepath.Join(root, "internal")
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid DataKit internal directory: %s", base)
	}
	var result []registryOverride
	fset := token.NewFileSet()
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		aliases := map[string]bool{}
		for _, imp := range file.Imports {
			name, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if name != "github.com/GuanceCloud/pipeline-go/ptinput/funcs" {
				continue
			}
			alias := "funcs"
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			if alias == "." {
				return fmt.Errorf("registry inventory cannot resolve dot import: %s", path)
			}
			aliases[alias] = true
		}
		if len(aliases) == 0 {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		isSelector := func(expr ast.Expr) (string, bool) {
			sel, ok := expr.(*ast.SelectorExpr)
			if !ok {
				return "", false
			}
			id, ok := sel.X.(*ast.Ident)
			return sel.Sel.Name, ok && aliases[id.Name]
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			add := func(node ast.Node, kind, registry, name string, dynamic bool) {
				result = append(result, registryOverride{location{filepath.ToSlash(rel), fset.Position(node.Pos()).Line}, fn.Name.Name, kind, registry, name, dynamic, "unmapped"})
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.AssignStmt:
					for _, lhs := range n.Lhs {
						if idx, ok := lhs.(*ast.IndexExpr); ok {
							if reg, ok := isSelector(idx.X); ok && (reg == "FuncsMap" || reg == "FuncsCheckMap") {
								name, known := literal(idx.Index)
								add(lhs, "entry_assignment", reg, name, !known)
							}
						} else if reg, ok := isSelector(lhs); ok && (reg == "FuncsMap" || reg == "FuncsCheckMap") {
							add(lhs, "registry_assignment", reg, "", false)
						}
					}
				case *ast.CallExpr:
					if name, ok := isSelector(n.Fun); ok && (name == "SetNetFilter" || name == "ReplaceNetFilter") {
						add(n, "network_policy", "", "http_request", false)
					}
					if id, ok := n.Fun.(*ast.Ident); ok && id.Name == "delete" && len(n.Args) == 2 {
						if reg, ok := isSelector(n.Args[0]); ok && (reg == "FuncsMap" || reg == "FuncsCheckMap") {
							name, known := literal(n.Args[1])
							add(n, "entry_delete", reg, name, !known)
						}
					}
				}
				return true
			})
		}
		return nil
	})
	return result, err
}
