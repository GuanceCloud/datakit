// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func ReplaceChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType)
	}
	return nil
}

func Replace(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func %s expects 3 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	var pattern, dz string

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		dz = funcExpr.Param[2].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType)
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("regular expression %s parse err: %w",
			reflect.TypeOf(funcExpr.Param[1]).String(), err)
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	newCont := reg.ReplaceAllString(cont, dz)
	if err := ctx.AddKey2PtWithVal(key, newCont, ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
