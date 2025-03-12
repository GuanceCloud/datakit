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

func StrfmtChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expects more than 2 args", funcExpr.Name), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param fmt expects StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}
	return nil
}

func Strfmt(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	outdata := make([]interface{}, 0)

	if len(funcExpr.Param) < 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected more than 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var fmts string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		fmts = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param fmt expect StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		v, _, _ := runtime.RunStmt(ctx, funcExpr.Param[i])
		outdata = append(outdata, v)
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	_ = addKey2PtWithVal(ctx.InData(), key, strfmt, ast.String, ptinput.KindPtDefault)

	return nil
}
