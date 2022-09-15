// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func TrimChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 1 || len(funcExpr.Param) > 2 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return fmt.Errorf("param type expect StringLiteral, got `%s'",
				funcExpr.Param[1].NodeType)
		}
	}
	return nil
}

func Trim(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
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
			cutset = funcExpr.Param[1].StringLiteral.Val
		default:
			return fmt.Errorf("param type expect StringLiteral, got `%s'",
				funcExpr.Param[1].NodeType)
		}
	}

	var val string
	if cutset == "" {
		val = strings.TrimSpace(cont)
	} else {
		val = strings.Trim(cont, cutset)
	}

	if err = ctx.AddKey2PtWithVal(key, val, ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
