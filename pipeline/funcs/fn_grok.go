// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func GrokChecking(ng *parser.EngineData, node parser.Node) error {
	// 只能在 checking 函数中修改 engine 的 grok
	g, ok := ng.GetEngineRGrok()
	if !ok {
		return fmt.Errorf("no grok obj")
	}

	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var pattern string
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	var re *grok.GrokRegexp
	var err error
	if ng.StackDeep() > 0 {
		deP := []map[string]*grok.GrokPattern{}
		deP = append(deP, g.GlobalDenormalizedPatterns, g.DenormalizedPatterns)
		deP = append(deP, ng.PatternStack()...)
		re, err = grok.CompilePattern(pattern, deP...)
		if err != nil {
			return err
		}
	} else {
		re, err = grok.CompilePattern(pattern, g.GlobalDenormalizedPatterns, g.DenormalizedPatterns)
		if err != nil {
			return err
		}
	}
	if re == nil {
		return fmt.Errorf("compile pattern `%s` failed", pattern)
	}
	if _, ok := g.CompliedGrokRe[pattern]; !ok {
		g.CompliedGrokRe[pattern] = make(map[string]*grok.GrokRegexp)
	}
	g.CompliedGrokRe[pattern][ng.PatternIndex()] = re

	funcExpr.Grok = re
	return nil
}

func Grok(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)

	grokRe := funcExpr.Grok
	if grokRe == nil {
		return fmt.Errorf("no grok obj")
	}
	var err error

	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	val, err := ng.GetContentStr(key)
	if err != nil {
		return nil
	}

	m, _, err := grokRe.RunWithTypeInfo(val)
	if err != nil {
		return nil
	}

	for k, v := range m {
		if err := ng.SetContent(k, v); err != nil {
			l.Debug(err)
			return nil
		}
	}
	return nil
}
