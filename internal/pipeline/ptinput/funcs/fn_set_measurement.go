// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func SetMeasurementChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param type expect BoolLiteral, got `%s'",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}

	return nil
}

func SetMeasurement(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return nil
	}
	if dtype == ast.String {
		if val, ok := val.(string); ok {
			_ = setMeasurement(ctx.InData(), val)
		}
	}

	if len(funcExpr.Param) == 2 &&
		funcExpr.Param[1].NodeType == ast.TypeBoolLiteral {
		if funcExpr.Param[1].BoolLiteral.Val {
			switch funcExpr.Param[0].NodeType { //nolint:exhaustive
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				if key, err := getKeyName(funcExpr.Param[0]); err == nil {
					deletePtKey(ctx.InData(), key)
				}
			}
		}
	}

	return nil
}
