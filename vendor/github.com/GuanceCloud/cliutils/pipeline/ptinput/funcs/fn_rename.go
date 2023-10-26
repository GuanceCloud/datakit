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

func RenameChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expect Identifier or AttrExpr, got `%s'",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}
	return nil
}

func Rename(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	to, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	from, err := getKeyName(funcExpr.Param[1])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
	}

	_ = renamePtKey(ctx.InData(), to, from)

	return nil
}
