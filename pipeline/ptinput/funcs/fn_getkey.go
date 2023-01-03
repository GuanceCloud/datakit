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

func GetkeyChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	return nil
}

func Getkey(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if funcExpr == nil {
		return runtime.NewRunError(ctx, "unreachable", funcExpr.NamePos)
	}
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf("func %s expected 1 args",
			funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if val, dtype, err := getPtKey(ctx.InData(), key); err != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
	} else {
		ctx.Regs.ReturnAppend(val, dtype)
	}

	return nil
}
