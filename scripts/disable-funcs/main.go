// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	bstoml "github.com/BurntSushi/toml"
)

type conf struct {
	Src       []string            `toml:"src"`
	SkipDirs  []string            `toml:"skip_dirs"`
	SkipFiles []string            `toml:"skip_files"`
	PkgFuncs  map[string][]string `toml:"pkg_functions"`

	scannedFiles, skippedFiles int
	bad                        []string
}

var (
	flagSrcDir   = flag.String("src", "", "Golang source code dir list")
	flagSkipDirs = flag.String("skip-dirs", "", "Skipped dirs list")
	skipDirs     []string

	flagConf = flag.String("config", "", "Configure file")
)

func (c *conf) skipped(path string) bool {
	for _, s := range c.SkipDirs {
		if filepath.Dir(path) == s {
			return true
		}
	}

	for _, s := range c.SkipFiles {
		if path == s {
			return true
		}
	}

	return false
}

func (c *conf) disabled(pkg, fn string) bool {
	if arr, ok := c.PkgFuncs[pkg]; ok {
		for _, f := range arr {
			if f == fn {
				return true
			}
		}
	}

	return false
}

func (c *conf) walk(dir string) {
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case ".go":
			c.scannedFiles++

			if c.skipped(path) {
				c.skippedFiles++
				return nil
			}

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if err != nil {
				log.Printf("failed to parse %q: %s, ignored", path, err)
				return nil
			}

			// 遍历 AST
			ast.Inspect(node, func(n ast.Node) bool {
				if callExpr, ok := n.(*ast.CallExpr); ok {
					if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						pkgIdent, ok := selExpr.X.(*ast.Ident)
						if ok && c.disabled(pkgIdent.Name, selExpr.Sel.Name) {
							c.bad = append(c.bad, fmt.Sprintf("Call of %s.%s (%s)\n",
								pkgIdent.Name, selExpr.Sel.Name, fset.Position(n.Pos())))
						}
					}
				}
				return true
			})
		default: // pass
		}

		return nil
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	c := conf{
		PkgFuncs: map[string][]string{},
	}

	if x, err := os.ReadFile(*flagConf); err != nil {
		log.Printf("[E] ReadFile(%q): %s", *flagConf, err)
		os.Exit(-1)
	} else {
		if _, err := bstoml.Decode(string(x), &c); err != nil {
			log.Printf("[E] unable to decode toml: %s", err)
			os.Exit(-1)
		}
	}

	for _, dir := range c.Src {
		log.Printf("[I] scanning dir %q...", dir)
		c.walk(dir)
	}

	if len(c.bad) > 0 {
		for _, x := range c.bad {
			log.Printf("[E] %s", x)
		}
		os.Exit(-1)
	}

	log.Printf("scanned %d files, %d skipped", c.scannedFiles, c.skippedFiles)
}
