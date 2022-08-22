package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	plrefertable "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/refertable"
)

func MQueryReferTableChecking(ngData *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)

	err := reIndexFuncArgs(funcExpr, []string{"table_name", "keys", "values"}, 3)
	if err != nil {
		return err
	}

	switch funcExpr.Param[0].(type) {
	case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects StringLiteral, Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case parser.FuncArgList:
		for _, v := range v {
			switch v.(type) {
			case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
			default:
				return fmt.Errorf("expect StringLiteral, Identifier or AttrExpr in FuncArgList, got %s",
					reflect.TypeOf(v).String())
			}
		}
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects FuncArgList, Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case parser.FuncArgList:
		for _, v := range v {
			switch v.(type) {
			case *parser.Identifier, *parser.AttrExpr,
				*parser.StringLiteral, *parser.NumberLiteral, *parser.BoolLiteral:
			default:
				return fmt.Errorf("expect Identifier, AttrExpr, "+
					"StringLiteral, NumberLiteral or BoolLiteral in FuncArgList, got %s",
					reflect.TypeOf(v).String())
			}
		}
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expects StringLiteral, NumberLiteral, BoolLiteral, AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	return nil
}

func MQueryReferTableMulti(ngData *parser.EngineData, node parser.Node) any {
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

	var colName []string

	switch v := funcExpr.Param[1].(type) {
	case parser.FuncArgList:
		for _, v := range v {
			switch v := v.(type) {
			case *parser.StringLiteral:
				colName = append(colName, v.Val)
			case *parser.Identifier, *parser.AttrExpr:
				val, err := ngData.GetContent(v)
				if err != nil {
					l.Error(err)
					return err
				}
				switch name := val.(type) {
				case string:
					colName = append(colName, name)
				default:
					err := fmt.Errorf("unsupported column name value type %s", reflect.TypeOf(val).String())
					l.Error(err)
					return err
				}
			}
		}
	case *parser.Identifier, *parser.AttrExpr:
		val, err := ngData.GetContent(v)
		if err != nil {
			l.Error(err)
			return err
		}
		switch cols := val.(type) {
		case []string:
			colName = cols
		default:
			err := fmt.Errorf("unsupported column name value type %s", reflect.TypeOf(val).String())
			l.Error(err)
			return err
		}
	}

	var colValue []any
	switch v := funcExpr.Param[2].(type) {
	case parser.FuncArgList:
		for _, v := range v {
			switch v := v.(type) {
			case *parser.StringLiteral:
				colValue = append(colValue, v.Value())
			case *parser.NumberLiteral:
				colValue = append(colValue, v.Value())
			case *parser.BoolLiteral:
				colValue = append(colValue, v.Value())
			case *parser.Identifier, *parser.AttrExpr:
				val, err := ngData.GetContent(v)
				if err != nil {
					l.Error(err)
					return err
				}
				colValue = append(colValue, val)
			}
		}
	case *parser.Identifier, *parser.AttrExpr:
		val, err := ngData.GetContent(v)
		if err != nil {
			l.Error(err)
			return err
		}
		switch cols := val.(type) {
		case []any:
			colValue = cols
		default:
			err := fmt.Errorf("unsupported column value type %s", reflect.TypeOf(val).String())
			l.Error(err)
			return err
		}
	}

	if vMap, ok := plrefertable.QueryReferTable(tableName,
		colName, colValue, nil); ok {
		for k, v := range vMap {
			_ = ngData.SetContent(k, v)
		}
	}
	return nil
}
