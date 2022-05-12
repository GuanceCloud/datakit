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

func AddPatternChecking(ngData *parser.EngineData, node parser.Node) error {
	g := ngData.GetGrok()

	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var name, pattern string
	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
		name = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	deep := ngData.StackDeep()
	pStack := ngData.PatternStack()
	if _, ok := g.DenormalizedPatterns[name]; ok && deep == 0 {
		return nil
		// return fmt.Errorf("pattern %s redefine", name)
	}
	dePatterns := []map[string]string{}
	dePatterns = append(dePatterns, g.GlobalDenormalizedPatterns, g.DenormalizedPatterns)
	dePatterns = append(dePatterns, pStack...)

	de, err := grok.DenormalizePattern(pattern, dePatterns...)
	if err != nil {
		return err
	}
	if deep < 0 {
		return fmt.Errorf("stack deep %d", deep)
	}
	if deep == 0 {
		g.DenormalizedPatterns[name] = de
	} else {
		pStack[deep-1][name] = de
	}

	return nil
}

func AddPattern(ng *parser.EngineData, node parser.Node) interface{} {
	return nil
}
