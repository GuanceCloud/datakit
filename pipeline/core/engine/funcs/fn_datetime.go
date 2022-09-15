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

func DateTimeChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("param `precision` expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType)
	}
	return nil
}

func DateTime(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	var precision, fmts string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		precision = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("param `precision` expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		fmts = funcExpr.Param[2].StringLiteral.Val
	default:
		return fmt.Errorf("param `fmt` expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType)
	}

	cont, err := ng.GetKey(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if v, err := DateFormatHandle(cont.Value, precision, fmts); err != nil {
		return err
	} else if err := ng.AddKey2PtWithVal(key, v, ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
