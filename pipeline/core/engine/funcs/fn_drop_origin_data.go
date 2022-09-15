// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func DropOriginDataChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	return nil
}

func DropOriginData(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	ctx.DeleteKeyPt("message")
	return nil
}
