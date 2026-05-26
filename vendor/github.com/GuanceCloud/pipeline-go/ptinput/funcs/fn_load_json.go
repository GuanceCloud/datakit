// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func LoadJSONChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	return nil
}

func LoadJSON(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[0].StartPos())
	}

	jsonStr := strings.TrimSpace(val.(string))
	if jsonStr == "" {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	if jsonStr[0] != '{' && jsonStr[0] != '[' && json.Valid([]byte(jsonStr)) {
		res := gjson.Parse(jsonStr)
		switch res.Type {
		case gjson.Number:
			ctx.Regs.ReturnAppend(res.Float(), ast.Float)
			return nil
		case gjson.True, gjson.False:
			ctx.Regs.ReturnAppend(res.Bool(), ast.Bool)
			return nil
		case gjson.String:
			ctx.Regs.ReturnAppend(res.String(), ast.String)
			return nil
		case gjson.Null:
			ctx.Regs.ReturnAppend(nil, ast.Nil)
			return nil
		}
	}

	var m any
	errJ := json.Unmarshal([]byte(jsonStr), &m)
	if errJ != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	switch v := m.(type) {
	case nil:
		ctx.Regs.ReturnAppend(nil, ast.Nil)
	case bool:
		ctx.Regs.ReturnAppend(v, ast.Bool)
	case float64:
		ctx.Regs.ReturnAppend(v, ast.Float)
	case string:
		ctx.Regs.ReturnAppend(v, ast.String)
	case []any:
		ctx.Regs.ReturnAppend(v, ast.List)
	case map[string]any:
		ctx.Regs.ReturnAppend(v, ast.Map)
	default:
		ctx.Regs.ReturnAppend(nil, ast.Nil)
	}
	return nil
}
