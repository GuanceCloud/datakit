package funcs

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func PtNameChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) > 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' can have at most one parameter", funcExpr.Name), funcExpr.NamePos)
	}

	return nil
}

func PtName(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) == 1 {
		if val, _, err := runtime.RunStmt(ctx, funcExpr.Param[0]); err == nil {
			if val, ok := val.(string); ok {
				_ = setPtName(ctx.InData(), val)
			}
		}
	}

	ctx.Regs.ReturnAppend(getPtName(ctx.InData()), ast.String)
	return nil
}
