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

func URLParseChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "prefix",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType), funcExpr.Param[0].StartPos())
	}
	if funcExpr.Param[1] != nil {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeIdentifier, ast.TypeStringLiteral, ast.TypeAttrExpr:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"expect StringLiteral or Identifier or AttrExpr, got %s",
				funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
		}
	}
	return nil
}

func URLParse(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType),
			funcExpr.Param[0].StartPos())
	}
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	prefix := ""
	if funcExpr.Param[1] != nil {
		prefixVal, dtype, errR := runtime.RunStmt(ctx, funcExpr.Param[1])
		if errR != nil {
			return errR
		}
		if dtype != ast.String {
			return runtime.NewRunError(ctx, "param data type expect string",
				funcExpr.Param[1].StartPos())
		}
		prefix = prefixVal.(string)
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
		prefix + "scheme": uu.Scheme,
		prefix + "host":   uu.Host,
		prefix + "port":   uu.Port(),
		prefix + "path":   uu.Path,
		prefix + "params": params,
	}
	ctx.Regs.ReturnAppend(res, ast.Map)
	return nil
}
