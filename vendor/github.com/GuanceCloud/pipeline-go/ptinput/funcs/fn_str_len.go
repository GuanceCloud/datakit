package funcs

import (
	"fmt"
	"unicode/utf8"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func StrLenChecking(ctx *runtime.Task, fn *ast.CallExpr) *errchain.PlError {
	if len(fn.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s` expect 1 arg", fn.Name), fn.NamePos)
	}
	return nil
}

func StrLen(ctx *runtime.Task, fn *ast.CallExpr) *errchain.PlError {
	val, dtype, err := runtime.RunStmt(ctx, fn.Param[0])
	if err != nil {
		return err
	}
	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			fn.Param[0].StartPos())
	}
	lenStr := utf8.RuneCountInString(val.(string))

	ctx.Regs.ReturnAppend(int64(lenStr), ast.Int)
	return nil
}
