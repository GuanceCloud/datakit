// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func DurationPrecisionChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	return nil
}

func DurationPrecision(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 3 {
		return fmt.Errorf("func `%s' expected 3 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}
	varb, err := ctx.GetKey(key)
	if err != nil {
		return err
	}
	var tValue int64
	if varb.DType != ast.Int {
		return fmt.Errorf("param value type expect int")
	} else {
		tValue = cast.ToInt64(varb.Value)
	}

	oldNew := [2]int{}
	for i := 1; i < 3; i += 1 {
		var err error
		switch funcExpr.Param[i].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			if oldNew[i-1], err = precision(funcExpr.Param[i].StringLiteral.Val); err != nil {
				l.Debug(err)
				return nil
			}
		default:
			return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
				reflect.TypeOf(funcExpr.Param[i]).String())
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
	if err := ctx.AddKey2PtWithVal(key, tValue, ast.Int, runtime.KindPtDefault); err != nil {
		return err
	}
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
