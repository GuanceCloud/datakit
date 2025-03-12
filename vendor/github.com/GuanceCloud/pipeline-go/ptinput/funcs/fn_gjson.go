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
	"github.com/tidwall/gjson"
)

func GJSONChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"input", "json_path", "key_name",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr, IndexExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	if funcExpr.Param[2] != nil {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeIdentifier, ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	return nil
}

func GJSON(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	srcKey, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	jpath, dtype, errP := runtime.RunStmt(ctx, funcExpr.Param[1])
	if errP != nil {
		return errP
	}
	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[1].StartPos())
	}

	targetKey := jpath.(string)

	if funcExpr.Param[2] != nil {
		tk, dtype, errP := runtime.RunStmt(ctx, funcExpr.Param[2])
		if errP != nil {
			return errP
		}
		if dtype != ast.String {
			return runtime.NewRunError(ctx, "param data type expect string",
				funcExpr.Param[2].StartPos())
		}

		targetKey = tk.(string)
	}

	cont, err := ctx.GetKeyConv2Str(srcKey)
	if err != nil {
		l.Debug(err)
		return nil
	}

	res := gjson.Get(cont, jpath.(string))
	rType := res.Type

	var v any
	switch rType {
	case gjson.Number:
		v = res.Float()
		dtype = ast.Float
	case gjson.True, gjson.False:
		v = res.Bool()
		dtype = ast.Bool
	case gjson.String, gjson.JSON:
		v = res.String()
		dtype = ast.String
	default:
		return nil
	}

	_ = addKey2PtWithVal(ctx.InData(), targetKey, v, dtype, ptinput.KindPtDefault)
	return nil
}
