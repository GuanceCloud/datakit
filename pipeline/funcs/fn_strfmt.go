package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func StrfmtChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func `%s' expects more than 2 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
	default:
		return fmt.Errorf("param fmt expects StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}
	return nil
}

func Strfmt(ng *parser.Engine, node parser.Node) error {
	outdata := make([]interface{}, 0)

	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func `%s' expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	var fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return fmt.Errorf("param fmt expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Identifier:
			data, _ := ng.GetContent(v)
			outdata = append(outdata, data)
		case *parser.AttrExpr:
			data, _ := ng.GetContent(v)
			outdata = append(outdata, data)
		case *parser.NumberLiteral:
			if v.IsInt {
				outdata = append(outdata, v.Int)
			} else {
				outdata = append(outdata, v.Float)
			}
		default:
			outdata = append(outdata, v)
		}
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	if err := ng.SetContent(key, strfmt); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
