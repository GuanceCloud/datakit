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

func DefaultTimeWithFmtChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected more than 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expect AttrExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	if len(funcExpr.Param) > 2 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param key expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	return nil
}

func DefaultTimeWithFmt(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var err error
	var goTimeFmt string
	var tz string
	var t time.Time
	timezone := time.Local

	if len(funcExpr.Param) < 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected more than 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		goTimeFmt = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param key expect StringLiteral, got %s",
			funcExpr.Param[1]), funcExpr.Param[1].StartPos())
	}

	if len(funcExpr.Param) > 2 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			tz = funcExpr.Param[2].StringLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param key expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	timeStr, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if tz != "" {
		timezone, err = time.LoadLocation(tz)
	}

	if err == nil {
		t, err = time.ParseInLocation(goTimeFmt, timeStr, timezone)
	}

	if err != nil {
		l.Debugf("time string: %s, time format: %s, timezone: %s, error msg: %s",
			timeStr, goTimeFmt, tz, err)
		return nil
	} else {
		_ = addKey2PtWithVal(ctx.InData(), key, t.UnixNano(), ast.Int, ptinput.KindPtDefault)
		return nil
	}
}
