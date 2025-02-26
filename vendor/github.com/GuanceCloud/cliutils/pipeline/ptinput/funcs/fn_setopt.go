// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var (
	_ = "fn setopt(status_mapping: bool = true)"
)

func mustAssignmentExpr(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if funcExpr == nil {
		return nil
	}

	for _, param := range funcExpr.Param {
		if param == nil {
			continue
		}
		if param.NodeType != ast.TypeAssignmentExpr {
			return runtime.NewRunError(ctx, "must be a named parameter", param.StartPos())
		}
	}
	return nil
}

func SetoptChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := mustAssignmentExpr(ctx, funcExpr); err != nil {
		return err
	}

	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"status_mapping",
	}, 0); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func Setopt(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) > 0 {
		param := funcExpr.Param[0]
		if param == nil {
			return nil
		}
		v, _, err := runtime.RunStmt(ctx, param)
		if err != nil {
			return err
		}
		if v != nil {
			if b, ok := v.(bool); ok {
				ptIn, errg := getPoint(ctx.InData())
				if errg != nil {
					return nil
				}
				ptIn.SetStatusMapping(b)
			}
		}
	}
	return nil
}
