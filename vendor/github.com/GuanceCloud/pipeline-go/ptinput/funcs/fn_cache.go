package funcs

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func CacheGetChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 1 arg", funcExpr.Name), funcExpr.NamePos)
	}
	return nil
}

func CacheGet(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if dtype != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[0].StartPos())
	}

	pt, errP := getPoint(ctx.InData())
	if errP != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	c := pt.GetCache()
	if c == nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	v, exist, errG := c.Get(val.(string))
	if !exist || errG != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	ctx.Regs.ReturnAppend(v, ast.String)

	return nil
}

func CacheSetChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "value", "exp",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func CacheSet(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	key, keyType, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	if keyType != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[0].StartPos())
	}

	value, _, err := runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}

	var expiration int64 = 100
	if funcExpr.Param[2] != nil {
		exp, expType, err := runtime.RunStmt(ctx, funcExpr.Param[2])
		if err != nil {
			return err
		}
		if expType != ast.Int {
			return runtime.NewRunError(ctx, "param data type expect int or nil",
				funcExpr.Param[2].StartPos())
		}
		expiration = exp.(int64)
	}

	pt, errP := getPoint(ctx.InData())
	if errP != nil {
		return nil
	}

	c := pt.GetCache()
	if c == nil {
		return nil
	}
	_ = c.Set(key.(string), value, time.Second*time.Duration(expiration))

	return nil
}
