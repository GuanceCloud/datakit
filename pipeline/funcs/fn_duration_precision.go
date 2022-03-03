package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func DurationPrecisionChecking(node parser.Node) error {
	return nil
}

func DurationPrecision(ng *parser.Engine, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func `%s' expected 3 args", funcExpr.Name)
	}

	var key parser.Node
	var tValue int64
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
		// if the key exists in the data
		if value, err := ng.GetContent(key); err == nil {
			if value, ok := value.(int64); ok {
				tValue = value
			} else {
				return fmt.Errorf("param value type expect int")
			}
		} else {
			return err
		}
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	oldNew := [2]int{}
	for i := 1; i < 3; i += 1 {
		var err error
		switch v := funcExpr.Param[i].(type) {
		case *parser.StringLiteral:
			if oldNew[i-1], err = precision(v.Val); err != nil {
				l.Debug(err)
				return nil
			}
		default:
			return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
				reflect.TypeOf(funcExpr.Param[i]).String())
		}
	}

	delta := oldNew[1] - oldNew[0]
	deltaAbs := delta
	if delta < 0 {
		deltaAbs = -delta
	}
	for i := 0; i < deltaAbs; i++ {
		if delta < 0 {
			tValue /= 10
		} else if delta > 0 {
			tValue *= 10
		}
	}
	if err := ng.SetContent(key, tValue); err != nil {
		return err
	}
	return nil
}

func precision(p string) (int, error) {
	switch strings.ToLower(p) {
	case "s":
		return 0, nil
	case "ms":
		return 3, nil
	case "us":
		return 6, nil
	case "ns":
		return 9, nil
	default:
		return 0, fmt.Errorf("unknow precision: %s", p)
	}
}
