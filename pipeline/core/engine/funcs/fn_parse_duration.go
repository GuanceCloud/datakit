// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func ParseDurationChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}
	return nil
}

func ParseDuration(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		l.Debugf("parse_duration(): invalid param")

		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	duStr, ok := cont.Value.(string)
	if !ok {
		return fmt.Errorf("parse_duration() expect string arg")
	}

	l.Debugf("parse duration %s", duStr)
	du, err := time.ParseDuration(duStr)
	if err != nil {
		l.Debug(err)
		return nil
	}

	if err := ctx.AddKey2PtWithVal(key, int64(du), ast.Int,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}
	return nil
}
