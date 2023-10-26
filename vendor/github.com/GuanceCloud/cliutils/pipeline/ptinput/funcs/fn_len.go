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

func LenChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1", funcExpr.Name), funcExpr.NamePos)
	}
	return nil
}

func Len(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	switch dtype { //nolint:exhaustive
	case ast.Map:
		ctx.Regs.ReturnAppend(int64(len(val.(map[string]any))), ast.Int)
	case ast.List:
		ctx.Regs.ReturnAppend(int64(len(val.([]any))), ast.Int)
	case ast.String:
		ctx.Regs.ReturnAppend(int64(len(val.(string))), ast.Int)
	default:
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
	}
	return nil
}
