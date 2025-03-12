// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/spf13/cast"
)

func QueryReferTableChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	err := normalizeFuncArgsDeprecated(funcExpr, []string{"table_name", "key", "value"}, 3)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral, ast.TypeIdentifier, ast.TypeAttrExpr:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expects StringLiteral, Identifier or AttrExpr, got %s",
			funcExpr.Param[1]), funcExpr.Param[1].StartPos())
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier, ast.TypeAttrExpr,
		ast.TypeStringLiteral, ast.TypeBoolLiteral,
		ast.TypeFloatLiteral, ast.TypeIntegerLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expects StringLiteral, NumberLiteral, BoolLiteral, AttrExpr or Identifier, got %s",
			funcExpr.Param[2]), funcExpr.Param[2].StartPos())
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

func QueryReferTable(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var tableName string

	tname, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param expect string", funcExpr.Param[0].StartPos())
	}

	tableName = tname.(string)

	var colName string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		colName = funcExpr.Param[1].StringLiteral().Val
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[1])
		val, err := ctx.GetKey(key)
		if err != nil {
			l.Debug(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
		switch val.DType { //nolint:exhaustive
		case ast.String:
			colName = cast.ToString(val.Value)
		default:
			err := fmt.Errorf("unsupported column name value type %s", val.DType)
			l.Debug(err)
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
	default:
		return nil
	}

	var colValue any
	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier, ast.TypeAttrExpr:
		key, _ := getKeyName(funcExpr.Param[2])
		if val, err := ctx.GetKey(key); err != nil {
			l.Debugf("key %s not found: %v", key, err)
			return nil
		} else {
			colValue = val.Value
		}
	case ast.TypeStringLiteral:
		colValue = funcExpr.Param[2].StringLiteral().Val
	case ast.TypeIntegerLiteral:
		colValue = funcExpr.Param[2].IntegerLiteral().Val
	case ast.TypeFloatLiteral:
		colValue = funcExpr.Param[2].FloatLiteral().Val
	case ast.TypeBoolLiteral:
		colValue = funcExpr.Param[2].BoolLiteral().Val
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

	pt, errR := getPoint(ctx.InData())
	if errR != nil {
		return nil
	}
	refT := pt.GetPlReferTables()
	if refT == nil {
		return nil
	}

	if vMap, ok := refT.Query(tableName,
		[]string{colName}, []any{colValue}, nil); ok {
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
