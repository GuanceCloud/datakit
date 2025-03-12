package funcs

import (
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var (
	defaultStreamTags = []string{"filepath", "host"}
)

func PtWindowChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	// fn pt_window(before, after, stream_tags := ["filepath", "host"])
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"before", "after", "stream_tags",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func PtWindow(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	before, vtype, errP := runtime.RunStmt(ctx, funcExpr.Param[0])
	if errP != nil {
		return errP
	}
	bf := before.(int64)

	if vtype != ast.Int {
		return runtime.NewRunError(ctx, "param data type expect int",
			funcExpr.Param[0].StartPos())
	}

	after, vtype, errP := runtime.RunStmt(ctx, funcExpr.Param[1])
	if errP != nil {
		return errP
	}
	af := after.(int64)

	if vtype != ast.Int {
		return runtime.NewRunError(ctx, "param data type expect int",
			funcExpr.Param[1].StartPos())
	}

	var tags []string
	if funcExpr.Param[2] != nil {
		streamTags, vtype, errP := runtime.RunStmt(ctx, funcExpr.Param[2])
		if errP != nil {
			return errP
		}
		if vtype != ast.List {
			return runtime.NewRunError(ctx, "param data type expect array",
				funcExpr.Param[2].StartPos())
		}
		if tagKey, ok := streamTags.([]any); ok && len(tagKey) != 0 {
			for _, v := range tagKey {
				if tag, ok := v.(string); ok {
					tags = append(tags, tag)
				} else {
					return nil
				}
			}
		}
	}
	if len(tags) == 0 {
		tags = defaultStreamTags
	}

	tagsVal := make([]string, 0, len(tags))
	for _, v := range tags {
		if val, err := ctx.GetKey(v); err != nil {
			return nil
		} else {
			if val.DType == ast.String {
				tagsVal = append(tagsVal, val.Value.(string))
			} else {
				return nil
			}
		}
	}

	pt, err := getPoint(ctx.InData())
	if err != nil {
		return nil
	}

	pt.PtWinRegister(int(bf), int(af), tags, tagsVal)
	return nil
}

func PtWindowHitChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	return nil
}

func PtWindowHit(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	pt, err := getPoint(ctx.InData())
	if err != nil {
		return nil
	}

	pt.PtWinHit()
	return nil
}
