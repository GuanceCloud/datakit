// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func ConvTraceIDW3C2DDChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	return nil
}

func ConvTraceIDW3C2DD(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 1 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	k, err := ctx.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	w3cTraceID, ok := k.Value.(string)
	if !ok {
		return nil
	}

	if ddTraceID, err := convTraceW3CToDD(w3cTraceID); err != nil {
		l.Debug(err)
	} else {
		addKey2PtWithVal(ctx.InData(), key, ddTraceID, ast.String, ptinput.KindPtDefault)
	}
	return nil
}

func convTraceW3CToDD(hexStr string) (string, error) {
	switch len(hexStr) {
	case 32:
		hexStr = hexStr[16:]
	case 16:
	default:
		return "", fmt.Errorf("not trace/span id")
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	i := binary.BigEndian.Uint64(b)
	return strconv.FormatUint(i, 10), nil
}
