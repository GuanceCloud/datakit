// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"regexp"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func MatchChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	param0 := funcExpr.Param[0]
	if param0.NodeType != ast.TypeStringLiteral {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, got %s",
			param0.NodeType), param0.StartPos())
	}

	// 预先编译正则表达式
	re, err := regexp.Compile(param0.StringLiteral.Val)
	if err != nil {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func match: regexp compile failed: %s", err.Error()), param0.StartPos())
	}
	funcExpr.Re = re

	param1 := funcExpr.Param[1]
	if !isPlVarbOrFunc(param1) && param1.NodeType != ast.TypeStringLiteral {
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteral, Identifier or AttrExpr, got %s",
			param1.NodeType), param1.StartPos())
	}

	return nil
}

func Match(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		err := fmt.Errorf("func %s expects 2 args", funcExpr.Name)
		l.Debug(err)
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	if funcExpr.Re == nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	cnt, err := getStr(ctx, funcExpr.Param[1])
	if err != nil {
		l.Debug(err)
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	isMatch := funcExpr.Re.MatchString(cnt)

	ctx.Regs.ReturnAppend(isMatch, ast.Bool)

	return nil
}
