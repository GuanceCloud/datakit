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

func SetTagChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral, ast.TypeIdentifier, ast.TypeAttrExpr:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param type expect StringLiteral, got `%s'",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}

	return nil
}

func SetTag(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 && len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 or 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var val any
	var dtype ast.DType
	if len(funcExpr.Param) == 2 {
		var errR *errchain.PlError
		// 不限制值的数据类型，如果不是 string 类将在设置为 tag 时自动转换为 string
		val, dtype, errR = runtime.RunStmt(ctx, funcExpr.Param[1])
		if errR != nil {
			return errR
		}
	} else {
		v, err := ctx.GetKey(key)
		if err != nil {
			l.Debug(err)
			_ = addKey2PtWithVal(ctx.InData(), key, "", ast.String, ptinput.KindPtTag)

			return nil
		}
		val = v.Value
		dtype = v.DType
	}

	_ = addKey2PtWithVal(ctx.InData(), key, val, dtype, ptinput.KindPtTag)

	return nil
}
