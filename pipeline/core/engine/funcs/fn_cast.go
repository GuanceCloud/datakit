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

func CastChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		switch funcExpr.Param[1].StringLiteral.Val {
		case "bool", "int", "float", "str", "string":
		default:
			return fmt.Errorf("unsupported data type: %s", funcExpr.Param[1].StringLiteral.Val)
		}
	default:
		return fmt.Errorf("param type expect StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType)
	}
	return nil
}

func Cast(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	var castType string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		castType = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("param type expect StringLiteral, got `%s'",
			funcExpr.Param[1].NodeType)
	}

	v, err := ctx.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	val, dtype := doCast(v.Value, castType)
	if err = ctx.AddKey2PtWithVal(key, val, dtype,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
