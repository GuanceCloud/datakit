// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func MatchChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}

	param0 := funcExpr.Param[0]
	if param0.NodeType != ast.TypeStringLiteral {
		return fmt.Errorf("expect StringLiteral, got %s",
			param0.NodeType)
	}

	// 预先编译正则表达式
	re, err := regexp.Compile(param0.StringLiteral.Val)
	if err != nil {
		return fmt.Errorf("func match: regexp compile failed: %w", err)
	}
	funcExpr.Re = re

	param1 := funcExpr.Param[1]
	if !isPlVarbOrFunc(param1) && param1.NodeType != ast.TypeStringLiteral {
		return fmt.Errorf("expect StringLiteral, Identifier or AttrExpr, got %s",
			param1.NodeType)
	}

	return nil
}

func Match(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 2 {
		err := fmt.Errorf("func %s expects 2 args", funcExpr.Name)
		l.Debug(err)
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}

	if funcExpr.Re == nil {
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}

	cnt, err := getStr(ctx, funcExpr.Param[1])
	if err != nil {
		l.Debug(err)
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}

	isMatch := funcExpr.Re.MatchString(cnt)

	ctx.Regs.Append(isMatch, ast.Bool)

	return nil
}
