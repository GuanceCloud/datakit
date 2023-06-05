// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func URLParseChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}
	return nil
}

func URLParse(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	u, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	uu, err := url.Parse(u)
	if err != nil {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"failed to parse url %s: %s", u, err.Error()), funcExpr.NamePos)
	}

	params := make(map[string]any)
	for k, vs := range uu.Query() {
		params[k] = strings.Join(vs, ",")
	}
	res := map[string]any{
		"scheme": uu.Scheme,
		"host":   uu.Host,
		"port":   uu.Port(),
		"path":   uu.Path,
		"params": params,
	}
	ctx.Regs.ReturnAppend(res, ast.Map)
	return nil
}
