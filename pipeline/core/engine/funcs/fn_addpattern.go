// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

func AddPatternChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var name, pattern string
	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		name = funcExpr.Param[0].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[0].NodeType)
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	deP, err := grok.DenormalizePattern(pattern, ctx)
	if err != nil {
		return err
	}
	ctx.SetPattern(name, deP)
	return nil
}

func AddPattern(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	return nil
}
