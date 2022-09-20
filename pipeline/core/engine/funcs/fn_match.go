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

func MatchChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 2 {
		return fmt.Errorf("func %s expected 2", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr , Identifier or StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}
	return nil
}

func Match(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var cont string
	var err error
	var res bool

	// var words, pattern parser.Node

	word, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err = ctx.GetKeyConv2Str(word)
	if err != nil {
		return err
	}

	var pattern string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
		p, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[1])
		if err != nil {
			return err
		}
		if dtype == ast.String {
			pattern, _ = p.(string)
		}
	default:
		l.Debugf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}

	res, err = isMatch(cont, pattern)
	if err != nil {
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}
	if res {
		ctx.Regs.Append(true, ast.Bool)
		return nil
	} else {
		ctx.Regs.Append(false, ast.Bool)
		return nil
	}
}

func isMatch(words string, pattern string) (bool, error) {
	isMatch, err := regexp.MatchString(pattern, "'"+words+"'")
	if err != nil {
		return false, err
	}
	return isMatch, err
}
