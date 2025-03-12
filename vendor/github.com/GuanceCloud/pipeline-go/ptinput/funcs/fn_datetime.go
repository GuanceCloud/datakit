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
	timefmt "github.com/itchyny/timefmt-go"
	conv "github.com/spf13/cast"
)

func DateTimeChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "precision", "fmt", "tz",
	}, 3); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param `precision` expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param `fmt` expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}

	if funcExpr.Param[3] != nil {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param `tz` expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[3].StartPos())
		}
	}

	// switch
	return nil
}

func DateTime(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 3 or 4 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var precision, fmts string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		precision = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param `precision` expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		fmts = funcExpr.Param[2].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param `fmt` expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	ts := conv.ToInt64(cont.Value)

	switch precision {
	case "us":
		ts *= 1e3
	case "ms":
		ts *= 1e6
	case "s":
		ts *= 1e9
	default:
	}

	t := time.Unix(0, ts)

	var tz string
	if funcExpr.Param[3] != nil {
		if funcExpr.Param[3].NodeType == ast.TypeStringLiteral {
			tz = funcExpr.Param[3].StringLiteral().Val
		} else {
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param `tz` expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[3].StartPos())
		}
	}

	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
		}
		t = t.In(loc)
	}

	if datetimeInnerFormat(fmts) {
		v, err := DateFormatHandle(&t, fmts)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
		}
		_ = addKey2PtWithVal(ctx.InData(), key, v, ast.String, ptinput.KindPtDefault)
	} else {
		_ = addKey2PtWithVal(ctx.InData(), key, timefmt.Format(t, fmts),
			ast.String, ptinput.KindPtDefault)
	}

	return nil
}
