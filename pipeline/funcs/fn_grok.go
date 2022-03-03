package funcs

import (
	"fmt"
	"reflect"

	vgrok "github.com/vjeantet/grok"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func CreateGrok(pattern map[string]string) (*vgrok.Grok, error) {
	return vgrok.NewWithConfig(&vgrok.Config{
		SkipDefaultPatterns: true,
		NamedCapturesOnly:   true,
		Patterns:            pattern,
	})
}

func GrokChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("expect Identifier or AttrExpr, got %s",
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

func Grok(ng *parser.Engine, node parser.Node) error {
	g := ng.GetGrok()
	var err error

	if g == nil {
		g, err = CreateGrok(ng.GetPatterns())
		if err != nil {
			l.Warn(err)
			return nil
		}
		ng.SetGrok(g)
	}

	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var pattern string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	val, err := ng.GetContentStr(key)
	if err != nil {
		l.Warn(err)
		return nil
	}

	m, err := g.Parse(pattern, val)
	if err != nil {
		l.Warn(err)
		return nil
	}

	for k, v := range m {
		err := ng.SetContent(k, v)
		if err != nil {
			l.Warn(err)
			return nil
		}
	}
	return nil
}
