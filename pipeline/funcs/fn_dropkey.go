package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DropkeyChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	return nil
}

func Dropkey(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	_ = ng.DeleteContent(key.String())

	return nil
}
