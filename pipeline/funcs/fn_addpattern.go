package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func AddPatternChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	return nil
}

func AddPattern(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if funcExpr.RunOk {
		return nil
	}

	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	defer func() {
		funcExpr.RunOk = true
	}()

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

	_ = ng.SetPatterns(map[string]string{name: pattern})

	return nil
}
