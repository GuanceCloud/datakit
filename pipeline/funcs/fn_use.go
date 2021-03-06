// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func UseChecking(ngData *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
		ngData.SetCallRef(v.Val, nil)
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	return nil
}

func Use(ngData *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	var refNg *parser.Engine
	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
		if ng, ok := ngData.GetCallRef(v.Val); ok {
			refNg = ng
		} else {
			l.Debugf("script not found: %s", v.Val)
			return nil
		}
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	return refNg.RefRun(ngData)
}
