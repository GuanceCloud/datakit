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

func LenChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expected 1", funcExpr.Name)
	}
	return nil
}

func Len(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	switch dtype { //nolint:exhaustive
	case ast.Map:
		ctx.Regs.Append(int64(len(val.(map[string]any))), ast.Int)
	case ast.List:
		ctx.Regs.Append(int64(len(val.([]any))), ast.Int)
	case ast.String:
		ctx.Regs.Append(int64(len(val.(string))), ast.Int)
	default:
		ctx.Regs.Append(int64(0), ast.Int)
	}
	return nil
}
