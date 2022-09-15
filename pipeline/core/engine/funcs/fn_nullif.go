// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func NullIfChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	return nil
}

func NullIf(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		// l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var val interface{}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral, ast.TypeNumberLiteral,
		ast.TypeBoolLiteral, ast.TypeNilLiteral:
		if v, _, err := runtime.RunStmt(ctx, funcExpr.Param[1]); err == nil {
			val = v
		}
	}

	// todo key string
	if reflect.DeepEqual(cont.Value, val) {
		ctx.DeleteKeyPt(key)
	}

	return nil
}
