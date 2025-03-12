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

func DefaultTimeChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s Expect at least one arg", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if len(funcExpr.Param) > 1 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param key expect StringLiteral, got %s",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}

	return nil
}

func DefaultTime(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expect at least one arg", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var tz string
	if len(funcExpr.Param) > 1 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			tz = funcExpr.Param[1].StringLiteral().Val
		default:
			err = fmt.Errorf("param key expect StringLiteral, got %s",
				funcExpr.Param[1].NodeType)
			usePointTime(ctx, key, err)
			return nil
		}
	}

	if nanots, err := TimestampHandle(cont, tz); err != nil {
		usePointTime(ctx, key, err)
	} else {
		addKey2PtWithVal(ctx.InData(), key, nanots, ast.Int, ptinput.KindPtDefault)
	}

	return nil
}

func usePointTime(ctx *runtime.Task, key string, err error) {
	_ = addKey2PtWithVal(ctx.InData(), runtime.PlRunInfoField, fmt.Sprintf("time convert failed: %v", err),
		ast.String, ptinput.KindPtDefault)

	_ = addKey2PtWithVal(ctx.InData(), key, pointTime(ctx.InData()), ast.Int, ptinput.KindPtDefault)
}
