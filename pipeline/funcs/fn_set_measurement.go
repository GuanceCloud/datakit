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

func SetMeasurementChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].(type) {
		case *parser.BoolLiteral:
		default:
			return fmt.Errorf("param type expect BoolLiteral, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	return nil
}

func SetMeasurement(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
		// if the key exists in the data
		if value, err := ng.GetContentStr(key); err == nil {
			ng.SetMeasurement(value)
		}
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if len(funcExpr.Param) == 2 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.BoolLiteral:
			if v.Val {
				_ = ng.DeleteContent(key)
			}
		default:
			return fmt.Errorf("param type expect StringLiteral, AttrExpr or Identifier, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}
	return nil
}
