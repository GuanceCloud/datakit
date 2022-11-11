// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func SampleChecking(_ *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeNumberLiteral && funcExpr.Param[0].NodeType != ast.TypeArithmeticExpr {
		return fmt.Errorf("expect NumberLiteral or ArithmeticExpr, got %s", funcExpr.Param[0].NodeType)
	}
	return nil
}

func Sample(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeNumberLiteral && funcExpr.Param[0].NodeType != ast.TypeArithmeticExpr {
		return fmt.Errorf("expect NumberLiteral or ArithmeticExpr, got %s", funcExpr.Param[0].NodeType)
	}

	var probability float64
	if funcExpr.Param[0].NodeType == ast.TypeArithmeticExpr {
		res, dtype, err := runtime.RunArithmeticExpr(ctx, funcExpr.Param[0].ArithmeticExpr)
		if err != nil {
			return err
		}
		if dtype != ast.Float && dtype != ast.Int {
			return fmt.Errorf("data type of the result of arithmetic expression is neither float nor int")
		}
		p, ok := res.(float64)
		if !ok {
			return fmt.Errorf("failed to convert expression evaluation result to float")
		}
		probability = p
	} else {
		if funcExpr.Param[0].NumberLiteral.IsInt {
			probability = float64(funcExpr.Param[0].NumberLiteral.Int)
		} else {
			probability = funcExpr.Param[0].NumberLiteral.Float
		}
	}

	if probability < 0 || probability > 1 {
		return fmt.Errorf("sampling probability should be in the range [0, 1]")
	}
	res := time.Now().UnixMicro()%100 <= int64(probability*100)
	ctx.Regs.Append(res, ast.Bool)
	return nil
}
