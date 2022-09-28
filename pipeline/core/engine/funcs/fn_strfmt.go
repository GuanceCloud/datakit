// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func StrfmtChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func `%s' expects more than 2 args", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("param fmt expects StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType)
	}
	return nil
}

func Strfmt(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	outdata := make([]interface{}, 0)

	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func `%s' expected more than 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	var fmts string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		fmts = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("param fmt expect StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType)
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		v, _, _ := runtime.RunStmt(ctx, funcExpr.Param[i])
		outdata = append(outdata, v)
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	if err := ctx.AddKey2PtWithVal(key, strfmt, ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
