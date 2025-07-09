// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func doAppendChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, "func %s expects 2 args", funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"the first param expects Identifier, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}

	return nil
}

func AppendChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	return doAppendChecking(ctx, funcExpr)
}

func Append(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if errR := doAppendChecking(ctx, funcExpr); errR != nil {
		return errR
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	val, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v` does not exist, ignored", key)
		return nil //nolint:nilerr
	}
	if val.DType != ast.List {
		l.Debugf("cannot append to a %s", val.DType.String())
		return nil
	}

	var arr []any
	switch v := val.Value.(type) {
	case []any:
		arr = v
	default:
		l.Debugf("expect []any, got %T", v)
		return nil
	}

	elem, _, errR := runtime.RunStmt(ctx, funcExpr.Param[1])
	if errR != nil {
		return errR
	}

	arr = append(arr, elem)
	ctx.Regs.ReturnAppend(arr, ast.List)
	return nil
}
