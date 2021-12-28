package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func SetTagChecking(node parser.Node) error {
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
		case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
		default:
			return fmt.Errorf("param type expect StringLiteral, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	return nil
}

func SetTag(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}

	tValue := ""

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
		// if the key exists in the data
		if value, err := ng.GetContentStr(key); err == nil {
			tValue = value
		}
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if len(funcExpr.Param) == 2 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.AttrExpr, *parser.Identifier:
			if value, err := ng.GetContentStr(v); err == nil {
				tValue = value
			} else {
				return err
			}
		case *parser.StringLiteral:
			tValue = v.Val
		default:
			return fmt.Errorf("param type expect StringLiteral, AttrExpr or Identifier, got `%s'",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}
	_ = ng.SetTag(key, tValue)

	return nil
}
