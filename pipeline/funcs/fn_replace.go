package funcs

import (
	"fmt"
	"reflect"
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func ReplaceChecking(node parser.Node) error {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}
	return nil
}

func Replace(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	var key parser.Node
	var pattern, dz string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		dz = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("regular expression %s parse err: %w",
			reflect.TypeOf(funcExpr.Param[1]).String(), err)
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	newCont := reg.ReplaceAllString(cont, dz)
	if err := ng.SetContent(key, newCont); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
