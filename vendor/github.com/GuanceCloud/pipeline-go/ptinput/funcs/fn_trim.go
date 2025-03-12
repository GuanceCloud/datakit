// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func TrimChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 1 || len(funcExpr.Param) > 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param type expect StringLiteral, got `%s'",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}
	return nil
}

func Trim(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	var cutset string
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			cutset = funcExpr.Param[1].StringLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param type expect StringLiteral, got `%s'",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}

	var val string
	if cutset == "" {
		val = strings.TrimSpace(cont)
	} else {
		val = strings.Trim(cont, cutset)
	}

	_ = addKey2PtWithVal(ctx.InData(), key, val, ast.String, ptinput.KindPtDefault)

	return nil
}
