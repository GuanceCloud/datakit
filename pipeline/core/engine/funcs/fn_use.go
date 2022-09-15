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

func UseChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		ctx.SetCallRef(funcExpr.Param[0].StringLiteral.Val)
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType)
	}

	return nil
}

func Use(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 args", funcExpr.Name)
	}

	var refScript *runtime.Script
	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		if ng, ok := ctx.GetCallRef(funcExpr.Param[0].StringLiteral.Val); ok {
			refScript = ng
		} else {
			l.Debugf("script not found: %s", funcExpr.Param[0].StringLiteral.Val)
			return nil
		}
	default:
		return fmt.Errorf("param key expects AttrExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType)
	}

	return runtime.RunScriptWithCtx(ctx, refScript)
}
