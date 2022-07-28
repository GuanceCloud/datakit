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

func NullIfChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	return nil
}

func NullIf(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	cont, err := ng.GetContent(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	// todo key string
	if reflect.DeepEqual(cont, val) {
		var k string

		switch t := key.(type) {
		case *parser.Identifier:
			k = t.String()
		case *parser.AttrExpr:
			k = t.String()
		case *parser.StringLiteral:
			k = t.Val
		default:
			l.Debugf("unsupported %v get", reflect.TypeOf(key).String())
			return nil
		}

		_ = ng.DeleteContent(k)
	}

	return nil
}
