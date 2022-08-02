// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func TrimChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 1 || len(funcExpr.Param) > 2 {
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
		case *parser.StringLiteral:
		default:
			return fmt.Errorf("param type expect StringLiteral, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}
	return nil
}

func Trim(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)
	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	var cutset string
	if len(funcExpr.Param) == 2 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.StringLiteral:
			cutset = v.Val
		default:
			return fmt.Errorf("param type expect StringLiteral, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	var val string
	if cutset == "" {
		val = strings.TrimSpace(cont)
	} else {
		val = strings.Trim(cont, cutset)
	}

	if err = ng.SetContent(key, val); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
