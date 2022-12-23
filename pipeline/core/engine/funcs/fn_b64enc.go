// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/base64"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func B64encChecking(_ *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return fmt.Errorf("expect Identifier, got %s", funcExpr.Param[0].NodeType)
	}
	return nil
}

func B64enc(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expects 1 arg", funcExpr.Name)
	}
	if funcExpr.Param[0].NodeType != ast.TypeIdentifier {
		return fmt.Errorf("expect Identifier, got %s", funcExpr.Param[0].NodeType)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}
	val, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v` does not exist, ignored", key)
		return nil //nolint:nilerr
	}
	var cont string

	switch v := val.Value.(type) {
	case string:
		cont = v
	default:
		l.Debugf("key `%s` has value not of type string, ignored", key)
		return nil
	}
	res := base64.StdEncoding.EncodeToString([]byte(cont))
	if err := ctx.AddKey2PtWithVal(key, res, ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}
	return nil
}
