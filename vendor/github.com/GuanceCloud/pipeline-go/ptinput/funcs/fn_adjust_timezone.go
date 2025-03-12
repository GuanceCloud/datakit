// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

const defaultMinuteDelta = int64(2)

func AdjustTimezoneChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 1 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 1 arg", funcExpr.Name),
			funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	return nil
}

func AdjustTimezone(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	// 默认允许 +2 分钟误差
	minuteAllow := defaultMinuteDelta
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeFloatLiteral:
			minuteAllow = int64(funcExpr.Param[1].FloatLiteral().Val)
		case ast.TypeIntegerLiteral:
			minuteAllow = funcExpr.Param[1].IntegerLiteral().Val
		default:
		}
	}

	if minuteAllow > 15 {
		minuteAllow = 15
	} else if minuteAllow < 0 {
		minuteAllow = 0
	}

	minuteAllow *= int64(time.Minute)

	logTS, err := ctx.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	switch cont := logTS.Value.(type) {
	case int64:
		cont = detectTimezone(cont, time.Now().UnixNano(), minuteAllow)
		_ = addKey2PtWithVal(ctx.InData(), key, cont, ast.Int, ptinput.KindPtDefault)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param value expect int64, got `%s`", reflect.TypeOf(cont)),
			funcExpr.Param[0].StartPos())
	}

	return nil
}

func detectTimezone(logTS, nowTS, minuteDuration int64) int64 {
	logTS -= (logTS/int64(time.Hour) - nowTS/int64(time.Hour)) * int64(time.Hour)

	if logTS-nowTS > minuteDuration {
		logTS -= int64(time.Hour)
	} else if logTS-nowTS <= -int64(time.Hour)+minuteDuration {
		logTS += int64(time.Hour)
	}

	return logTS
}
