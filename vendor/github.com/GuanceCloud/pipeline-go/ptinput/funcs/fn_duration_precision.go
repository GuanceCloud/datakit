// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/spf13/cast"
)

func DurationPrecisionChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	return nil
}

func DurationPrecision(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 3 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 3 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	varb, err := ctx.GetKey(key)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}
	var tValue int64
	if varb.DType != ast.Int {
		return runtime.NewRunError(ctx, "param value type expect int",
			funcExpr.Param[0].StartPos())
	} else {
		tValue = cast.ToInt64(varb.Value)
	}

	oldNew := [2]int{}
	for i := 1; i < 3; i += 1 {
		var err error
		switch funcExpr.Param[i].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			if oldNew[i-1], err = precision(funcExpr.Param[i].StringLiteral().Val); err != nil {
				l.Debug(err)
				return nil
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"param key expect Identifier or AttrExpr, got `%s'",
				reflect.TypeOf(funcExpr.Param[i]).String()), funcExpr.Param[i].StartPos())
		}
	}

	delta := oldNew[1] - oldNew[0]
	deltaAbs := delta
	if delta < 0 {
		deltaAbs = -delta
	}
	for i := 0; i < deltaAbs; i++ {
		if delta < 0 {
			tValue /= 10
		} else if delta > 0 {
			tValue *= 10
		}
	}
	_ = addKey2PtWithVal(ctx.InData(), key, tValue, ast.Int,
		ptinput.KindPtDefault)
	return nil
}

func precision(p string) (int, error) {
	switch strings.ToLower(p) {
	case "s":
		return 0, nil
	case "ms":
		return 3, nil
	case "us":
		return 6, nil
	case "ns":
		return 9, nil
	default:
		return 0, fmt.Errorf("unknow precision: %s", p)
	}
}
