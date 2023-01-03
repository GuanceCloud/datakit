// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func AddPatternChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	var name, pattern string
	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		name = funcExpr.Param[0].StringLiteral.Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteral, got %s",
			funcExpr.Param[0].NodeType), funcExpr.NamePos)
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral.Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	deP, err := grok.DenormalizePattern(pattern, ctx)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	ctx.SetPattern(name, deP)
	return nil
}

func AddPattern(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	return nil
}
