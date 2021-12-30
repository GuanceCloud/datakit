package funcs

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DzChecking(node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}

	set := arglistForIndexOne(funcExpr)

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var start, end int

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else if v.IsInt {
		start = int(v.Int)
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = int(v.Int)
		}

		if start > end {
			return fmt.Errorf("range value start %v must be less than end %v", start, end)
		}
	}
	return nil
}

func Dz(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}

	set := arglistForIndexOne(funcExpr)

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var start, end int

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else if v.IsInt {
		start = int(v.Int)
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = int(v.Int)
		}

		if start > end {
			return fmt.Errorf("range value start %v must be less than end %v", start, end)
		}
	}

	cont, err := ng.GetContentStr(key)
	if err != nil {
		l.Debug("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if end > utf8.RuneCountInString(cont) {
		end = utf8.RuneCountInString(cont)
	}

	if start <= 0 {
		start = 1
	}

	arrCont := []rune(cont)

	for i := 0; i < len(arrCont); i++ {
		if i+1 >= start && i < end {
			if unicode.Is(unicode.Han, arrCont[i]) {
				arrCont[i] = rune('＊')
			} else {
				arrCont[i] = rune('*')
			}
		}
	}

	if err := ng.SetContent(key, string(arrCont)); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}
