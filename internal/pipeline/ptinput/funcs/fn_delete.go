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
	"github.com/spf13/cast"
)

func DeleteMapItemChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{"src", "key"}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeIndexExpr, ast.TypeIdentifier:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expect IndexExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	return nil
}

func DeleteMapItem(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	var keyName string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		keyName = funcExpr.Param[1].StringLiteral.Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeIndexExpr:
		indexExprDel(ctx, funcExpr.Param[0].IndexExpr, keyName)
		return nil
	case ast.TypeIdentifier:
		key, err := getKeyName(funcExpr.Param[0])
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
		}
		varb, err := ctx.GetKey(key)
		if err != nil {
			l.Debugf("key `%s' not exist, ignored", key)
			return nil
		}

		if varb.DType == ast.Map {
			switch v := varb.Value.(type) {
			case map[string]any: // for json map
				delete(v, keyName)
				return nil
			default:
				return nil
			}
		}
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param key expect IndexExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}

	return nil
}

func indexExprDel(ctx *runtime.Context, expr *ast.IndexExpr, keyDel string) {
	key := expr.Obj.Name

	varb, err := ctx.GetKey(key)
	if err != nil {
		return
	}

	switch varb.DType { //nolint:exhaustive
	case ast.List:
		switch varb.Value.(type) {
		case []any:
		default:
			return
		}
	case ast.Map:
		switch varb.Value.(type) {
		case map[string]any: // for json map
		default:
			return
		}
	default:
		return
	}

	searchListAndMap(ctx, varb.Value, expr.Index, keyDel)
}

func searchListAndMap(ctx *runtime.Context, obj any, index []*ast.Node, keyDel string) {
	cur := obj

	for _, i := range index {
		key, keyType, err := runtime.RunStmt(ctx, i)
		if err != nil {
			return
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if keyType != ast.String {
				return
			}
			var ok bool
			cur, ok = curVal[key.(string)]
			if !ok {
				return
			}
		case []any:
			if keyType != ast.Int {
				return
			}
			keyInt := cast.ToInt(key)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return
			}
			cur = curVal[keyInt]
		default:
			return
		}
	}

	if v, ok := cur.(map[string]any); ok {
		delete(v, keyDel)
	}
}
