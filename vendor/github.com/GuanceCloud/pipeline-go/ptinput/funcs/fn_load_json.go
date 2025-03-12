// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func LoadJSONChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	return nil
}

func LoadJSON(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	var m any

	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[0].StartPos())
	}
	errJ := json.Unmarshal([]byte(val.(string)), &m)
	if errJ != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}
	m, dtype = ast.DectDataType(m)

	ctx.Regs.ReturnAppend(m, dtype)
	return nil
}
