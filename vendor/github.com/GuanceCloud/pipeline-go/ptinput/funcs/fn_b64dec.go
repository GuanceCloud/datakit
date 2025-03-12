// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/base64"
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func B64decChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}
	return nil
}

func B64dec(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
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
	var cont string

	switch v := val.Value.(type) {
	case string:
		cont = v
	default:
		l.Debugf("key `%s` has value not of type string, ignored", key)
		return nil
	}
	res, err := base64.StdEncoding.DecodeString(cont)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	_ = addKey2PtWithVal(ctx.InData(), key, string(res), ast.String, ptinput.KindPtDefault)
	return nil
}
