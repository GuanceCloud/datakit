// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

func GrokChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if funcExpr.Grok != nil {
		return nil
	}

	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 3 {
		return fmt.Errorf("func %s expected 2 or 3 args", funcExpr.Name)
	}

	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return fmt.Errorf("param key expect BoolLiteral, got `%s'",
				funcExpr.Param[2].NodeType)
		}
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	var pattern string
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		pattern = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	gRe, err := grok.CompilePattern(pattern, ctx)
	if err != nil {
		return err
	}
	funcExpr.Grok = gRe
	return nil
}

func Grok(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	grokRe := funcExpr.Grok
	if grokRe == nil {
		ng.Regs.Append(false, ast.Bool)
		return fmt.Errorf("no grok obj")
	}
	var err error

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		ng.Regs.Append(false, ast.Bool)
		return err
	}

	val, err := ng.GetKeyConv2Str(key)
	if err != nil {
		ng.Regs.Append(false, ast.Bool)
		return nil
	}

	trimSpace := true
	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			trimSpace = funcExpr.Param[2].BoolLiteral.Val
		default:
			ng.Regs.Append(false, ast.Bool)
			return fmt.Errorf("param key expect BoolLiteral, got `%s'",
				funcExpr.Param[2].NodeType)
		}
	}

	m, _, err := grokRe.RunWithTypeInfo(val, trimSpace)
	if err != nil {
		ng.Regs.Append(false, ast.Bool)
		return nil
	}

	for k, v := range m {
		var dtype ast.DType
		if v == nil {
			dtype = ast.Nil
		} else {
			switch v.(type) {
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
		if err := ng.AddKey2PtWithVal(k, v, dtype, runtime.KindPtDefault); err != nil {
			l.Debug(err)
			ng.Regs.Append(false, ast.Bool)
			return nil
		}
	}
	ng.Regs.Append(true, ast.Bool)
	return nil
}
