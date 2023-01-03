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
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/parser"
)

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

		if err := CheckScript(p, check); err != nil {
			// TODO
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

func RunScriptWithoutMapIn(proc *plruntime.Script, data plruntime.InputWithoutMap, signal plruntime.Signal) *errchain.PlError {
	return plruntime.RunScriptWithoutMapIn(proc, data, signal)
}

func RunScriptWithRMapIn(proc *plruntime.Script, data plruntime.InputWithRMap, signal plruntime.Signal) *errchain.PlError {
	return plruntime.RunScriptWithRMapIn(proc, data, signal)
}

func RunScriptRef(ctx *plruntime.Context, proc *plruntime.Script) *errchain.PlError {
	return plruntime.RefRunScript(ctx, proc)
}

func CheckScript(proc *plruntime.Script, funcsCheck map[string]plruntime.FuncCheck) *errchain.PlError {
	return plruntime.CheckScript(proc, funcsCheck)
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
