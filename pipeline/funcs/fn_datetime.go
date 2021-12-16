package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DateTimeChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param `key` expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("param `precision` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}
	return nil
}

func DateTime(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}

	var key parser.Node
	var precision, fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param `key` expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		precision = v.Val
	default:
		return fmt.Errorf("param `precision` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	cont, err := ng.GetContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if v, err := DateFormatHandle(cont, precision, fmts); err != nil {
		return err
	} else if err := ng.SetContent(key, v); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
