// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"strconv"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ParseIntChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"val", "base",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func ParseInt(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	val, _, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if val == nil {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	valStr, ok := val.(string)
	if !ok {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	base, _, err := runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}

	if base == nil {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	baseInt, ok := base.(int64)
	if !ok {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	v, _ := strconv.ParseInt(valStr, int(baseInt), 64)
	ctx.Regs.ReturnAppend(v, ast.Int)
	return nil
}

func FormatIntChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"val", "base",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func FormatInt(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	val, _, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if val == nil {
		ctx.Regs.ReturnAppend("", ast.Int)
		return nil
	}

	valStr, ok := val.(int64)
	if !ok {
		ctx.Regs.ReturnAppend("", ast.Int)
		return nil
	}

	base, _, err := runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}

	if base == nil {
		ctx.Regs.ReturnAppend("", ast.Int)
		return nil
	}

	baseInt, ok := base.(int64)
	if !ok {
		ctx.Regs.ReturnAppend("", ast.Int)
		return nil
	}

	v := strconv.FormatInt(valStr, int(baseInt))
	ctx.Regs.ReturnAppend(v, ast.String)
	return nil
}
