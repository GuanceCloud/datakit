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

func SetMeasurementChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return fmt.Errorf("param type expect BoolLiteral, got `%s'",
				funcExpr.Param[1].NodeType)
		}
	}

	return nil
}

func SetMeasurement(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}

	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return nil
	}
	if dtype == ast.String {
		if val, ok := val.(string); ok {
			ctx.SetMeasurement(val)
		}
	}

	if len(funcExpr.Param) == 2 &&
		funcExpr.Param[1].NodeType == ast.TypeBoolLiteral {
		if funcExpr.Param[1].BoolLiteral.Val {
			switch funcExpr.Param[0].NodeType { //nolint:exhaustive
			case ast.TypeIdentifier, ast.TypeAttrExpr:
				if key, err := getKeyName(funcExpr.Param[0]); err == nil {
					ctx.DeleteKeyPt(key)
				}
			}
		}
	}

	return nil
}
