package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DefaultTimeChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 1 {
		return fmt.Errorf("func %s expected more than 1 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if len(funcExpr.Param) > 1 {
		switch funcExpr.Param[1].(type) {
		case *parser.StringLiteral:
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	return nil
}

func DefaultTime(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 1 {
		return fmt.Errorf("func %s expected more than 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var tz string
	if len(funcExpr.Param) > 1 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.StringLiteral:
			tz = v.Val
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[1]).String())
		}
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if v, err := TimestampHandle(cont, tz); err != nil {
		return fmt.Errorf("time convert fail error %w", err)
	} else if err := ng.SetContent(key, v); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
