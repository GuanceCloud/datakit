// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
	plrefertable "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/refertable"
)

func MQueryReferTableChecking(ngData *runtime.Context, funcExpr *ast.CallExpr) error {
	err := reIndexFuncArgs(funcExpr, []string{"table_name", "keys", "values"}, 3)
	if err != nil {
		return err
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeListInitExpr:
		for _, v := range funcExpr.Param[1].ListInitExpr.List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral, ast.TypeIdentifier, ast.TypeAttrExpr:
			default:
				return fmt.Errorf("expect StringLiteral, Identifier or AttrExpr in FuncArgList, got %s",
					reflect.TypeOf(v).String())
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
	default:
		return fmt.Errorf("param key expects FuncArgList, Identifier or AttrExpr, got %s",
			funcExpr.Param[1].NodeType)
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeListInitExpr:
		for _, v := range funcExpr.Param[2].ListInitExpr.List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeIdentifier, ast.TypeAttrExpr,
				ast.TypeStringLiteral, ast.TypeNumberLiteral, ast.TypeBoolLiteral:
			default:
				return fmt.Errorf("expect Identifier, AttrExpr, "+
					"StringLiteral, NumberLiteral or BoolLiteral in FuncArgList, got %s",
					v.NodeType)
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
	default:
		return fmt.Errorf("param key expects StringLiteral, NumberLiteral, BoolLiteral, AttrExpr or Identifier, got %s",
			funcExpr.Param[2].NodeType)
	}

	return nil
}

func MQueryReferTableMulti(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var tableName string

	tname, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	if dtype != ast.String {
		return fmt.Errorf("param expect string")
	}

	tableName = tname.(string)

	var colName []string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeListInitExpr:
		for _, v := range funcExpr.Param[1].ListInitExpr.List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral:
				colName = append(colName, v.StringLiteral.Val)
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				key, _ := getKeyName(v)
				val, err := ctx.GetKey(key)
				if err != nil {
					l.Error(err)
					return err
				}
				switch val.DType { //nolint:exhaustive
				case ast.String:
					colName = append(colName, cast.ToString(val.Value))
				default:
					err := fmt.Errorf("unsupported column name value type %s", val.DType)
					l.Error(err)
					return err
				}
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[1])

		val, err := ctx.GetKey(key)
		if err != nil {
			l.Error(err)
			return err
		}
		switch val.DType { //nolint:exhaustive
		case ast.String:
			colName = append(colName, cast.ToString(val.Value))
		default:
			err := fmt.Errorf("unsupported column name value type %s", val.DType)
			l.Error(err)
			return err
		}
	}

	var colValue []any
	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeListInitExpr:
		for _, v := range funcExpr.Param[2].ListInitExpr.List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral:
				colValue = append(colValue, v.StringLiteral.Val)
			case ast.TypeNumberLiteral:
				colValue = append(colValue, v.NumberLiteral.Value())
			case ast.TypeBoolLiteral:
				colValue = append(colValue, v.BoolLiteral.Val)
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				key, _ := getKeyName(v)
				val, err := ctx.GetKey(key)
				if err != nil {
					l.Debug(err)
					return err
				}
				colValue = append(colValue, val.Value)
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[2])
		val, err := ctx.GetKey(key)
		if err != nil {
			l.Error(err)
			return err
		}
		switch val.DType { //nolint:exhaustive
		case ast.List:
			if colval, ok := val.Value.([]any); ok {
				colValue = colval
			}
		default:
			err := fmt.Errorf("unsupported column value type %s", reflect.TypeOf(val).String())
			l.Error(err)
			return err
		}
	}

	if vMap, ok := plrefertable.QueryReferTable(tableName,
		colName, colValue, nil); ok {
		for k, v := range vMap {
			var dtype ast.DType
			switch v.(type) {
			case string:
				dtype = ast.String
			case bool:
				dtype = ast.Bool
			case int64:
				dtype = ast.Int
			case float64:
				dtype = ast.Float
			default:
				return nil
			}
			_ = ctx.AddKey2PtWithVal(k, v, dtype, runtime.KindPtDefault)
		}
	}
	return nil
}
