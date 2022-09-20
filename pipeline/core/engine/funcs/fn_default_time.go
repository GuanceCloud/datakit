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

func DefaultTimeChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 1 {
		return fmt.Errorf("func %s expected more than 1 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	if len(funcExpr.Param) > 1 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				funcExpr.Param[1].NodeType)
		}
	}

	return nil
}

func DefaultTime(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) < 1 {
		return fmt.Errorf("func %s expected more than 1 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
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
			tz = funcExpr.Param[1].StringLiteral.Val
		default:
			err = fmt.Errorf("param key expect StringLiteral, got %s",
				funcExpr.Param[1].NodeType)
			usePointTime(ctx, key, err)
			return nil
		}
	}

	if v, err := TimestampHandle(cont, tz); err != nil {
		usePointTime(ctx, key, err)
		return nil
	} else if err := ctx.AddKey2PtWithVal(key, v, ast.Int, runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}

func usePointTime(ctx *runtime.Context, key string, err error) {
	_ = ctx.AddKey2PtWithVal("pl_msg", fmt.Sprintf("time convert failed: %v", err),
		ast.String, runtime.KindPtDefault)
	_ = ctx.AddKey2PtWithVal(key, ctx.PointTime(), ast.Int, runtime.KindPtDefault)
}
