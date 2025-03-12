// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/spf13/cast"
)

func MQueryReferTableChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	err := normalizeFuncArgsDeprecated(funcExpr, []string{"table_name", "keys", "values"}, 3)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeListLiteral:
		for _, v := range funcExpr.Param[1].ListLiteral().List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral, ast.TypeIdentifier, ast.TypeAttrExpr:
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"expect StringLiteral, Identifier or AttrExpr in FuncArgList, got %s",
					reflect.TypeOf(v).String()), funcExpr.Param[1].StartPos())
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expects FuncArgList, Identifier or AttrExpr, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeListLiteral:
		for _, v := range funcExpr.Param[2].ListLiteral().List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeIdentifier, ast.TypeAttrExpr,
				ast.TypeStringLiteral, ast.TypeBoolLiteral,
				ast.TypeFloatLiteral, ast.TypeIntegerLiteral:
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"expect Identifier, AttrExpr, "+
						"StringLiteral, NumberLiteral or BoolLiteral in FuncArgList, got %s",
					v.NodeType), funcExpr.Param[2].StartPos())
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expects StringLiteral, NumberLiteral, BoolLiteral, AttrExpr or Identifier, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}

	return nil
}

func MQueryReferTableMulti(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var tableName string

	tname, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param expect string",
			funcExpr.Param[0].StartPos())
	}

	tableName = tname.(string)

	var colName []string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeListLiteral:
		for _, v := range funcExpr.Param[1].ListLiteral().List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral:
				colName = append(colName, v.StringLiteral().Val)
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				key, _ := getKeyName(v)
				val, err := ctx.GetKey(key)
				if err != nil {
					l.Error(err)
					return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
				}
				switch val.DType { //nolint:exhaustive
				case ast.String:
					colName = append(colName, cast.ToString(val.Value))
				default:
					err := fmt.Errorf("unsupported column name value type %s", val.DType)
					l.Error(err)
					return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
				}
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[1])

		val, err := ctx.GetKey(key)
		if err != nil {
			l.Error(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
		switch val.DType { //nolint:exhaustive
		case ast.String:
			colName = append(colName, cast.ToString(val.Value))
		default:
			err := fmt.Errorf("unsupported column name value type %s", val.DType)
			l.Error(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
	}

	var colValue []any
	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeListLiteral:
		for _, v := range funcExpr.Param[2].ListLiteral().List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral:
				colValue = append(colValue, v.StringLiteral().Val)
			case ast.TypeFloatLiteral:
				colValue = append(colValue, v.FloatLiteral().Val)
			case ast.TypeIntegerLiteral:
				colValue = append(colValue, v.IntegerLiteral().Val)
			case ast.TypeBoolLiteral:
				colValue = append(colValue, v.BoolLiteral().Val)
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				key, _ := getKeyName(v)
				val, err := ctx.GetKey(key)
				if err != nil {
					l.Debug(err)
					return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
				}
				colValue = append(colValue, val.Value)
			}
		}
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[2])
		val, err := ctx.GetKey(key)
		if err != nil {
			l.Error(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
		}
		switch val.DType { //nolint:exhaustive
		case ast.List:
			if colval, ok := val.Value.([]any); ok {
				colValue = colval
			}
		default:
			err := fmt.Errorf("unsupported column value type %s", reflect.TypeOf(val).String())
			l.Error(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
		}
	}

	pt, errR := getPoint(ctx.InData())
	if errR != nil {
		return nil
	}
	refT := pt.GetPlReferTables()
	if refT == nil {
		return nil
	}

	if vMap, ok := refT.Query(tableName,
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
			_ = addKey2PtWithVal(ctx.InData(), k, v, dtype, ptinput.KindPtDefault)
		}
	}
	return nil
}
