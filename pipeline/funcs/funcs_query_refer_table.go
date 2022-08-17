package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	plrefertable "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/refertable"
)

func QueryReferTableChecking(ngData *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)

	err := reIndexFuncArgs(funcExpr, []string{"table_name", "key", "value"}, 3)
	if err != nil {
		return err
	}

	switch funcExpr.Param[0].(type) {
	case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects StringLiteral, Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects StringLiteral, Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch funcExpr.Param[2].(type) {
	case *parser.Identifier, *parser.AttrExpr,
		*parser.StringLiteral, *parser.NumberLiteral, *parser.BoolLiteral:
	default:
		return fmt.Errorf("param key expects StringLiteral, NumberLiteral, BoolLiteral, AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	// TODO: pos param 4: selected([]string)

	// if len(funcExpr.Param) == 4 {
	// 	switch v := funcExpr.Param[3].(type) {
	// 	case parser.FuncArgList:
	// 		for _, item := range v {
	// 			switch item.(type) {
	// 			case *parser.StringLiteral:
	// 			default:
	// 				return fmt.Errorf("param key expects StringLiteral, got %s",
	// 					reflect.TypeOf(funcExpr.Param[2]).String())
	// 			}
	// 		}
	// 	case nil:
	// 	default:
	// 		return fmt.Errorf("param key expects FuncArgList, got %s",
	// 			reflect.TypeOf(funcExpr.Param[3]).String())
	// 	}
	// }

	return nil
}

func QueryReferTable(ngData *parser.EngineData, node parser.Node) any {
	funcExpr := fexpr(node)

	var tableName string
	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
		tableName = v.Val
	case *parser.Identifier, *parser.AttrExpr:
		val, err := ngData.GetContent(v)
		if err != nil {
			l.Debug(err)
			return err
		}
		switch name := val.(type) {
		case string:
			tableName = name
		default:
			err := fmt.Errorf("unsupported table param value type: %s", reflect.TypeOf(val).String())
			l.Debug(err)
			return err
		}
	default:
	}

	var colName string
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		colName = v.Val
	case *parser.Identifier, *parser.AttrExpr:
		val, err := ngData.GetContent(v)
		if err != nil {
			l.Debug(err)
			return err
		}
		switch name := val.(type) {
		case string:
			colName = name
		default:
			err := fmt.Errorf("unsupported column name value type %s", reflect.TypeOf(val).String())
			l.Debug(err)
			return err
		}
	default:
	}

	var colValue any
	switch v := funcExpr.Param[2].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		if val, err := ngData.GetContent(v); err != nil {
			l.Debugf("key %s not found: %w", v.String(), err)
			return nil
		} else {
			colValue = val
		}
	case *parser.StringLiteral:
		colValue = v.Value()
	case *parser.NumberLiteral:
		colValue = v.Value()
	case *parser.BoolLiteral:
		colValue = v.Value()
	}

	// TODO: pos param 4: selected([]string)

	// var selectd []string
	// switch v := funcExpr.Param[3].(type) {
	// case parser.FuncArgList:
	// 	for _, item := range v {
	// 		if item, ok := item.(*parser.StringLiteral); ok {
	// 			selectd = append(selectd, item.Val)
	// 		}
	// 	}
	// case nil:
	// }

	if vMap, ok := plrefertable.QueryReferTable(tableName,
		[]string{colName}, []any{colValue}, nil); ok {
		for k, v := range vMap {
			_ = ngData.SetContent(k, v)
		}
	}
	return nil
}
