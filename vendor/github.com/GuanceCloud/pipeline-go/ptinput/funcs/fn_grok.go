// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func GrokChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if funcExpr.Grok != nil {
		return nil
	}

	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 2 or 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param key expect BoolLiteral, got `%s'",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var pattern string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	gRe, err := grok.CompilePattern(pattern, ctx)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	funcExpr.Grok = gRe
	return nil
}

func Grok(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	grokRe := funcExpr.Grok
	if grokRe == nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return runtime.NewRunError(ctx, "no grok obj", funcExpr.NamePos)
	}
	var err error

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	val, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	trimSpace := true
	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			trimSpace = funcExpr.Param[2].BoolLiteral().Val
		default:
			ctx.Regs.ReturnAppend(false, ast.Bool)
			return runtime.NewRunError(ctx, fmt.Sprintf("param key expect BoolLiteral, got `%s'",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	if grokRe.WithTypeInfo() {
		result, err := grokRe.RunWithTypeInfo(val, trimSpace)
		if err != nil {
			ctx.Regs.ReturnAppend(false, ast.Bool)
			return nil
		}
		for i, name := range grokRe.MatchNames() {
			var dtype ast.DType
			if result[i] == nil {
				dtype = ast.Nil
			} else {
				switch result[i].(type) {
				case int64:
					dtype = ast.Int
				case float64:
					dtype = ast.Float
				case string:
					dtype = ast.String
				case bool:
					dtype = ast.Bool
				default:
					continue
				}
			}
			if ok := addKey2PtWithVal(ctx.InData(), name, result[i], dtype, ptinput.KindPtDefault); !ok {
				ctx.Regs.ReturnAppend(false, ast.Bool)
				return nil
			}
		}
	} else {
		result, err := grokRe.Run(val, trimSpace)
		if err != nil {
			ctx.Regs.ReturnAppend(false, ast.Bool)
			return nil
		}
		for i, name := range grokRe.MatchNames() {
			if ok := addKey2PtWithVal(ctx.InData(), name, result[i], ast.String, ptinput.KindPtDefault); !ok {
				ctx.Regs.ReturnAppend(false, ast.Bool)
				return nil
			}
		}
	}

	ctx.Regs.ReturnAppend(true, ast.Bool)
	return nil
}
