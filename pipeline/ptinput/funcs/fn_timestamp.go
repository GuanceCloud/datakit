// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func TimestampChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	err := reindexFuncArgs(funcExpr, []string{"precision"}, 0)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	return nil
}

func Timestamp(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	var precision string
	if funcExpr.Param[0] != nil {
		v, _, err := runtime.RunStmt(ctx, funcExpr.Param[0])
		if err != nil {
			return err
		}
		if v, ok := v.(string); ok {
			precision = v
		}
	}
	var ts int64
	switch precision {
	case "us":
		ts = time.Now().UnixMicro()
	case "ms":
		ts = time.Now().UnixMilli()
	case "s":
		ts = time.Now().Unix()
	default:
		ts = time.Now().UnixNano()
	}
	ctx.Regs.ReturnAppend(ts, ast.Int)
	return nil
}
