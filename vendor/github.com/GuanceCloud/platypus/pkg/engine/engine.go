// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package engine run pipeline script
package engine

import (
	"os"
	"path/filepath"

	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/engine/runtimev2"
	"github.com/GuanceCloud/platypus/pkg/parser"
)

func ParseV2(name, script string, fn map[string]*runtimev2.Fn) (*runtimev2.Script, error) {
	stmts, err := parser.ParsePipeline(name, script)
	if err != nil {
		return nil, err
	}

	p := &runtimev2.Script{
		Name:  name,
		Stmts: stmts,
		Fn:    fn,
	}

	if err := p.Check(); err != nil {
		return nil, err
	}

	return p, nil
}

func ParseScript(scripts map[string]string,
	call map[string]plruntime.FuncCall, check map[string]plruntime.FuncCheck) (
	map[string]*plruntime.Script, map[string]error,
) {
	retErrMap := map[string]error{}
	retMap := map[string]*plruntime.Script{}

	for name, content := range scripts {
		stmts, err := parser.ParsePipeline(name, content)
		if err != nil {
			retErrMap[name] = err
			continue
		}
		p := &plruntime.Script{
			FuncCall: call,
			Name:     name,
			Content:  content,
			Ast:      stmts,
		}

		if err := p.Check(check); err != nil {
			retErrMap[name] = err
			continue
		}
		retMap[name] = p
	}
	retMap, retErrs := EngineCallRefLinkAndCheck(retMap, retErrMap)

	for k, v := range retErrs {
		retErrMap[k] = v
	}

	return retMap, retErrMap
}

func ReadPlScriptFromDir(dirPath string) (map[string]string, map[string]string, error) {
	ret := map[string]string{}
	retPath := map[string]string{}
	dirPath = filepath.Clean(dirPath)
	if dirEntry, err := os.ReadDir(dirPath); err != nil {
		return nil, nil, err
	} else {
		for _, v := range dirEntry {
			if v.IsDir() {
				continue
			}
			sName := v.Name()
			if filepath.Ext(sName) != ".ppl" && filepath.Ext(sName) != ".p" {
				continue
			}
			sPath := filepath.Join(dirPath, sName)
			if name, script, err := ReadPlScriptFromFile(sPath); err == nil {
				ret[name] = script
				retPath[name] = sPath
			} else {
				return nil, nil, err
			}
		}
	}
	return ret, retPath, nil
}

func ReadPlScriptFromFile(fp string) (string, string, error) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(filepath.Clean(fp)); err == nil {
		_, sName := filepath.Split(fp)
		return sName, string(v), nil
	} else {
		return "", "", err
	}
}
