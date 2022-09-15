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

func AddkeyChecking(_ *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) > 2 || len(funcExpr.Param) < 1 {
		return fmt.Errorf("func %s expected 1 or 2 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	return nil
}

func AddKey(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if funcExpr == nil {
		return fmt.Errorf("unreachable")
	}
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expected 1 or 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	if len(funcExpr.Param) == 1 {
		_ = ctx.AddKey2Pt(key, runtime.KindPtDefault)
		return nil
	}

	var val any
	var dtype ast.DType

	val, dtype, err = runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}

	if err := ctx.AddKey2PtWithVal(key, val, dtype, runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
