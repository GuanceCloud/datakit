// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ValueTypeChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"val",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func ValueType(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if _, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0]); err != nil {
		ctx.Regs.ReturnAppend("", ast.String)
		return nil
	} else {
		var v string
		switch dtype { //nolint:exhaustive
		case ast.Bool:
			v = "bool"
		case ast.Int:
			v = "int"
		case ast.Float:
			v = "float"
		case ast.String:
			v = "str"
		case ast.List:
			v = "list"
		case ast.Map:
			v = "map"
		}
		ctx.Regs.ReturnAppend(v, ast.String)
		return nil
	}
}
