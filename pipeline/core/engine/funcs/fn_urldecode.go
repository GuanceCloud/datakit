// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func URLDecodeChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}
	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}
	return nil
}

func URLDecode(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if v, err := UrldecodeHandle(cont); err != nil {
		return err
	} else if err := ctx.AddKey2PtWithVal(key, v, ast.String, runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
