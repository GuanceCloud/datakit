// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/json"
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ValidJSONChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	return nil
}

func ValidJSON(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	val, _, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if val != nil {
		if v, ok := val.(string); ok {
			valid := json.Valid([]byte(v))
			ctx.Regs.ReturnAppend(valid, ast.Bool)
			return nil
		}
	}

	ctx.Regs.ReturnAppend(false, ast.Bool)
	return nil
}
