package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func GroupInChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral:
		default:
			return fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}
	return nil
}

func GroupIn(ng *parser.Engine, node parser.Node) error {
	setdata := make([]interface{}, 0)
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	set := arglistForIndexOne(funcExpr)
	value := funcExpr.Param[2]

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var newkey parser.Node
	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral, *parser.Identifier:
			newkey = v
		default:
			return fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	for _, node := range set {
		switch v := node.(type) {
		case *parser.Identifier:
			cont, err := ng.GetContent(v.Name)
			if err != nil {
				l.Debug("key `%v' not exist, ignored", key)
				return nil //nolint:nilerr
			}
			setdata = append(setdata, cont)
		case *parser.NumberLiteral:
			if v.IsInt {
				setdata = append(setdata, v.Int)
			} else {
				setdata = append(setdata, v.Float)
			}
		case *parser.BoolLiteral:
			setdata = append(setdata, v.Val)
		case *parser.StringLiteral:
			setdata = append(setdata, v.Val)
		default:
			setdata = append(setdata, v)
		}
	}

	cont, err := ng.GetContent(key)
	if err != nil {
		l.Debug("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if GroupInHandle(cont, setdata) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				if err := ng.SetContent(newkey, v.IsInt); err != nil {
					l.Warn(err)
					return nil
				}
			} else if err := ng.SetContent(newkey, v.Float); err != nil {
				l.Warn(err)
				return nil
			}
		case *parser.StringLiteral:
			if err := ng.SetContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return nil
			}
		case *parser.BoolLiteral:
			if err := ng.SetContent(newkey, v.Val); err != nil {
				l.Warn(err)
				return nil
			}
		}
	}

	return nil
}
