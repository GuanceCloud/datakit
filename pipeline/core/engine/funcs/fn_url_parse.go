// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"net/url"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func URLParseChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return fmt.Errorf("expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType)
	}
	return nil
}

func URLParse(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier && funcExpr.Param[0].NodeType != ast.TypeAttrExpr {
		return fmt.Errorf("expect Identifier or AttrExpr, got %s", funcExpr.Param[0].NodeType)
	}
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}
	u, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	uu, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", u, err)
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
	ctx.Regs.Append(res, ast.Map)
	return nil
}
