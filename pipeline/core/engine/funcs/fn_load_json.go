// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func LoadJSONChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func %s expected 1", funcExpr.Name)
	}
	return nil
}

func LoadJSON(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	val, dtype, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	var m any

	if dtype != ast.String {
		return fmt.Errorf("param data type expect string")
	}
	err = json.Unmarshal([]byte(val.(string)), &m)
	if err != nil {
		return err
	}
	m, dtype = ast.DectDataType(m)

	ctx.Regs.Append(m, dtype)
	return nil
}
