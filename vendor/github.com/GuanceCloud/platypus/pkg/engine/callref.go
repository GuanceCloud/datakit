// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package engine

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
)

type searchPath struct {
	nodeMap map[string]struct{}
	path    []string
}

func (p *searchPath) Push(nodeName string) error {
	p.path = append(p.path, nodeName)
	if _, ok := p.nodeMap[nodeName]; ok {
		defer func() {
			p.path = p.path[:len(p.path)-1]
		}()
		return fmt.Errorf("circular dependency: %s", p)
	}
	p.nodeMap[nodeName] = struct{}{}
	return nil
}

func (p *searchPath) Pop() {
	if len(p.path) == 0 {
		return
	}

	nodeName := p.path[len(p.path)-1]

	p.path = p.path[:len(p.path)-1]
	delete(p.nodeMap, nodeName)
}

func newSearchPath() *searchPath {
	return &searchPath{
		nodeMap: map[string]struct{}{},
		path:    []string{},
	}
}

func (p *searchPath) String() string {
	if len(p.path) == 0 {
		return ""
	}

	return strings.Join(p.path, " -> ")
}

type param struct {
	name    string
	namePos token.LnColPos

	allNg    map[string]*runtime.Script
	retMap   map[string]*runtime.Script
	allErrNg map[string]error
}

func EngineCallRefLinkAndCheck(allNg map[string]*runtime.Script, allErrNg map[string]error) (map[string]*runtime.Script, map[string]error) {
	retMap := map[string]*runtime.Script{}
	retErrMap := map[string]error{}

	for name, proc := range allNg {
		p := &param{
			name:     name,
			namePos:  token.InvalidLnColPos,
			allNg:    allNg,
			allErrNg: allErrNg,
			retMap:   retMap,
		}

		sPath := newSearchPath()
		if err := dfs(name, proc, sPath, p); err != nil {
			retErrMap[name] = err
		} else {
			retMap[name] = proc
		}
	}

	return retMap, retErrMap
}

func dfs(name string, procc *runtime.Script, sPath *searchPath, p *param) error {
	if err := sPath.Push(name); err != nil {
		return errchain.NewErr(p.name, p.namePos, err.Error())
	}

	if _, ok := p.retMap[name]; ok {
		return nil
	}

	for _, expr := range procc.CallRef {
		cName, err := getParamRefScript(expr)
		p.namePos = expr.NamePos
		if err != nil {
			return err
		}

		if cNg, ok := p.allNg[cName]; !ok {
			if err, ok := p.allErrNg[cName]; ok {
				if e, ok := err.(*errchain.PlError); ok {
					return e.Copy().ChainAppend(
						procc.Name, p.namePos)
				}
				return err
			}
			return errchain.NewErr(procc.Name, p.namePos,
				fmt.Sprintf("script %s not found", cName))
		} else {
			expr.PrivateData = cNg
			if err := dfs(cName, cNg, sPath, p); err != nil {
				if e, ok := err.(*errchain.PlError); ok {
					return e.Copy().ChainAppend(procc.Name, p.namePos)
				}
				return err
			}
		}
	}

	p.retMap[name] = procc
	sPath.Pop()

	return nil
}

func getParamRefScript(expr *ast.CallExpr) (string, error) {
	if expr == nil {
		return "", fmt.Errorf("nil ptr")
	}
	if expr.Name != "use" {
		return "", fmt.Errorf("function name is not 'use'")
	}
	if len(expr.Param) != 1 {
		return "", fmt.Errorf("the number of parameters is not 1")
	}

	if expr.Param[0].NodeType != ast.TypeStringLiteral {
		return "", fmt.Errorf("param type expects StringLiteral got `%s`", expr.Param[0].NodeType)
	}

	return expr.Param[0].StringLiteral.Val, nil
}
