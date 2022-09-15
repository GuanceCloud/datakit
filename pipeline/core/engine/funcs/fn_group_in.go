// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func GroupInChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeStringLiteral, ast.TypeIdentifier:
		default:
			return fmt.Errorf("param new-key expect AttrExpr, StringLiteral or Identifier, got %s",
				funcExpr.Param[3].NodeType)
		}
	}
	return nil
}

func GroupIn(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		// l.Debugf("key '%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var set []any
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		if v, err := ctx.GetKey(funcExpr.Param[1].Identifier.Name); err == nil &&
			v.DType == ast.List {
			if v, ok := v.Value.([]any); ok {
				set = v
			}
		}
	case ast.TypeListInitExpr:
		if v, dtype, err := runtime.RunListInitExpr(ctx,
			funcExpr.Param[1].ListInitExpr); err == nil && dtype == ast.List {
			if v, ok := v.([]any); ok {
				set = v
			}
		}
	}

	if GroupInHandle(cont.Value, set) {
		value := funcExpr.Param[2]

		var val any
		var dtype ast.DType

		switch value.NodeType { //nolint:exhaustive
		case ast.TypeNumberLiteral:
			if value.NumberLiteral.IsInt {
				val = value.NumberLiteral.Int
				dtype = ast.Int
			} else {
				val = value.NumberLiteral.Float
				dtype = ast.Float
			}
		case ast.TypeStringLiteral:
			val = value.StringLiteral.Val
			dtype = ast.String
		case ast.TypeBoolLiteral:
			val = value.BoolLiteral.Val
			dtype = ast.String
		default:
			return nil
		}

		if len(funcExpr.Param) == 4 {
			if k, err := getKeyName(funcExpr.Param[3]); err == nil {
				key = k
			} else {
				return err
			}
		}
		_ = ctx.AddKey2PtWithVal(key, val, dtype, runtime.KindPtDefault)
	}

	return nil
}
