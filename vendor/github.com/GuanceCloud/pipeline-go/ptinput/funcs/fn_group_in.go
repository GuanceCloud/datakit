// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func GroupInChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 3 or 4 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeStringLiteral, ast.TypeIdentifier:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param new-key expect AttrExpr, StringLiteral or Identifier, got %s",
				funcExpr.Param[3].NodeType), funcExpr.Param[3].StartPos())
		}
	}
	return nil
}

func GroupIn(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return runtime.NewRunError(ctx, fmt.Sprintf("func %s expected 3 or 4 args", funcExpr.Name),
			funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		// l.Debugf("key '%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var set []any
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		if v, err := ctx.GetKey(funcExpr.Param[1].Identifier().Name); err == nil &&
			v.DType == ast.List {
			if v, ok := v.Value.([]any); ok {
				set = v
			}
		}
	case ast.TypeListLiteral:
		if v, dtype, err := runtime.RunListInitExpr(ctx,
			funcExpr.Param[1].ListLiteral()); err == nil && dtype == ast.List {
			if v, ok := v.([]any); ok {
				set = v
			}
		}
	}

	if GroupInHandle(cont.Value, set) {
		value := funcExpr.Param[2]

		var val any
		var dtype ast.DType

		switch value.NodeType { //nolint:exhaustive
		case ast.TypeIntegerLiteral:
			val = value.IntegerLiteral().Val
			dtype = ast.Int
		case ast.TypeFloatLiteral:
			val = value.FloatLiteral().Val
			dtype = ast.Float
		case ast.TypeStringLiteral:
			val = value.StringLiteral().Val
			dtype = ast.String
		case ast.TypeBoolLiteral:
			val = value.BoolLiteral().Val
			dtype = ast.String
		default:
			return nil
		}

		if len(funcExpr.Param) == 4 {
			if k, err := getKeyName(funcExpr.Param[3]); err == nil {
				key = k
			} else {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[3].StartPos())
			}
		}
		_ = addKey2PtWithVal(ctx.InData(), key, val, dtype, ptinput.KindPtDefault)
	}

	return nil
}
