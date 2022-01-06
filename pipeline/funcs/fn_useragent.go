package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func UserAgentChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expects 1 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	return nil
}

func UserAgent(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expects 1 args", funcExpr.Name)
	}

	var key parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Debug("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	dic := UserAgentHandle(cont)

	for k, val := range dic {
		if err := ng.SetContent(k, val); err != nil {
			l.Warn(err)
			return nil
		}
	}

	return nil
}
