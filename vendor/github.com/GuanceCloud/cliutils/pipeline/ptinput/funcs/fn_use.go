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

func UseChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 args", funcExpr.Name), funcExpr.NamePos)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		ctx.SetCallRef(funcExpr)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expects StringLiteral, got %s",
			funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}

	return nil
}

func Use(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 args", funcExpr.Name), funcExpr.NamePos)
	}

	var refScript *runtime.Script
	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		if funcExpr.PrivateData != nil {
			if s, ok := funcExpr.PrivateData.(*runtime.Script); ok {
				refScript = s
			} else {
				l.Debugf("unknown error: %s", funcExpr.Param[0].StringLiteral.Val)
				return nil
			}
		} else {
			l.Debugf("script not found: %s", funcExpr.Param[0].StringLiteral.Val)
			return nil
		}
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expects StringLiteral, got %s",
			funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}

	if err := runtime.RefRunScript(ctx, refScript); err != nil {
		return err.ChainAppend(ctx.Name(), funcExpr.NamePos)
	}
	return nil
}
