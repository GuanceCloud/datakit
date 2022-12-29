// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func SampleChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral, ast.TypeFloatLiteral, ast.TypeArithmeticExpr:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect NumberLiteral or ArithmeticExpr, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}

	return nil
}

func Sample(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}

	var probability float64

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeArithmeticExpr:
		res, dtype, err := runtime.RunArithmeticExpr(ctx, funcExpr.Param[0].ArithmeticExpr)
		if err != nil {
			return err
		}
		if dtype != ast.Float && dtype != ast.Int {
			return runtime.NewRunError(ctx,
				"data type of the result of arithmetic expression is neither float nor int",
				funcExpr.Param[0].StartPos())
		}
		p, ok := res.(float64)
		if !ok {
			return runtime.NewRunError(ctx, "failed to convert expression evaluation result to float",
				funcExpr.Param[0].StartPos())
		}
		probability = p

	case ast.TypeFloatLiteral:
		probability = funcExpr.Param[0].FloatLiteral.Val
	case ast.TypeIntegerLiteral:
		probability = float64(funcExpr.Param[0].IntegerLiteral.Val)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect NumberLiteral or ArithmeticExpr, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}

	if probability < 0 || probability > 1 {
		l.Error("sampling probability should be in the range [0, 1]")
		return nil
		// return runtime.NewRunError(ctx,
		// 	"sampling probability should be in the range [0, 1]", funcExpr.NamePos)
	}

	res := time.Now().UnixMicro()%100 <= int64(probability*100)
	ctx.Regs.ReturnAppend(res, ast.Bool)
	return nil
}
