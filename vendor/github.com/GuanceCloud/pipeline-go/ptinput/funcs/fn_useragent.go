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

func UserAgentChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "prefix",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
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

func UserAgent(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
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

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	dic, dicT := UserAgentHandle(cont)

	for k, val := range dic {
		if dtype, ok := dicT[k]; ok {
			_ = addKey2PtWithVal(ctx.InData(), prefix+k, val, dtype, ptinput.KindPtDefault)
		}
	}

	return nil
}
