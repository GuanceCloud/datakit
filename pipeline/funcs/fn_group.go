package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func GroupChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
	}

	set := arglistForIndexOne(funcExpr)

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var start, end float64

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral, *parser.Identifier:
		default:
			return fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			start = float64(v.Int)
		} else {
			start = v.Float
		}
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = float64(v.Int)
		} else {
			end = v.Float
		}

		if start > end {
			return fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}
	return nil
}

func Group(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
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

	newkey := key
	var start, end float64

	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.StringLiteral, *parser.Identifier:
			newkey = v
		default:
			return fmt.Errorf("param new-key expect AttrExpr or StringLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			start = float64(v.Int)
		} else {
			start = v.Float
		}
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = float64(v.Int)
		} else {
			end = v.Float
		}

		if start > end {
			return fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}

	cont, err := ng.GetContent(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if GroupHandle(cont, start, end) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				if err := ng.SetContent(newkey, v.Int); err != nil {
					l.Warn(err)
					return nil
				}
			} else {
				if err := ng.SetContent(newkey, v.Float); err != nil {
					l.Warn(err)
					return nil
				}
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

		default:
			l.Errorf("unknown group elements: %s", reflect.TypeOf(value).String())
			return fmt.Errorf("unsupported group type")
		}
	}

	return nil
}
