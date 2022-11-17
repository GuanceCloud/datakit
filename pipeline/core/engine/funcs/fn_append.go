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

func doAppendChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return fmt.Errorf("the first param expects Identifier, got %s", funcExpr.Param[0].NodeType)
	}
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier, ast.TypeAttrExpr, ast.TypeBoolLiteral, ast.TypeNumberLiteral, ast.TypeStringLiteral:
	default:
		return fmt.Errorf("the second param expects Identifier, AttrExpr, BoolLiteral, NumberLiteral or StringLiteral, got %s", funcExpr.Param[1].NodeType) //nolint:lll
	}
	return nil
}

func AppendChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	return doAppendChecking(ctx, funcExpr)
}

func Append(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if err := doAppendChecking(ctx, funcExpr); err != nil {
		return err
	}

	elem, _, err := runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
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
	arr = append(arr, elem)
	ctx.Regs.Append(arr, ast.List)
	return nil
}
