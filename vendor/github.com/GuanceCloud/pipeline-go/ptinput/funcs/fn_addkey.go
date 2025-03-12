// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func AddkeyChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) > 2 || len(funcExpr.Param) < 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	return nil
}

func AddKey(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if len(funcExpr.Param) == 1 {
		v, err := ctx.GetKey(key)
		if err != nil {
			l.Debug(err)
			return nil
		}
		_ = addKey2PtWithVal(ctx.InData(), key, v.Value, v.DType, ptinput.KindPtDefault)
		return nil
	}

	var val any
	var dtype ast.DType

	val, dtype, errRun := runtime.RunStmt(ctx, funcExpr.Param[1])
	if errRun != nil {
		return errRun.ChainAppend(ctx.Name(), funcExpr.NamePos)
	}
	_ = addKey2PtWithVal(ctx.InData(), key, val, dtype, ptinput.KindPtDefault)

	return nil
}
