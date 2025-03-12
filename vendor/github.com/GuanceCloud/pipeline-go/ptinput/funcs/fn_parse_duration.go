// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ParseDurationChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	return nil
}

func ParseDuration(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		l.Debugf("parse_duration(): invalid param")

		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	duStr, ok := cont.Value.(string)
	if !ok {
		l.Debug(err)
		return nil
		// return runtime.NewRunError(ctx, "parse_duration() expect string arg",
		// 	funcExpr.Param[0].StartPos())
	}

	l.Debugf("parse duration %s", duStr)
	du, err := time.ParseDuration(duStr)
	if err != nil {
		l.Debug(err)
		return nil
	}

	_ = addKey2PtWithVal(ctx.InData(), key, int64(du), ast.Int, ptinput.KindPtDefault)
	return nil
}
