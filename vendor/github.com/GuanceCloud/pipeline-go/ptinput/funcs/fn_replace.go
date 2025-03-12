// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ReplaceChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var pattern string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return runtime.NewRunError(ctx, fmt.Sprintf("regular expression %s parse err: %s",
			reflect.TypeOf(funcExpr.Param[1]).String(), err.Error()), funcExpr.Param[1].StartPos())
	}

	funcExpr.PrivateData = reg

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}
	return nil
}

func Replace(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	reg, ok := funcExpr.PrivateData.(*regexp.Regexp)
	if !ok {
		return runtime.NewRunError(ctx, "regexp obj not found", funcExpr.Param[1].StartPos())
	}

	if len(funcExpr.Param) != 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var dz string

	switch funcExpr.Param[2].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		dz = funcExpr.Param[2].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect StringLiteral, got %s",
			funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	newCont := reg.ReplaceAllString(cont, dz)
	_ = addKey2PtWithVal(ctx.InData(), key, newCont, ast.String, ptinput.KindPtDefault)

	return nil
}
