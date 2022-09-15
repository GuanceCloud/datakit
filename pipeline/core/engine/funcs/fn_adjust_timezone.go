// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

const defaultMinuteDelta = int64(2)

func AdjustTimezoneChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 arg", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	return nil
}

func AdjustTimezone(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	// 默认允许 +2 分钟误差
	minuteAllow := defaultMinuteDelta
	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeNumberLiteral:
			if funcExpr.Param[1].NumberLiteral.IsInt {
				minuteAllow = funcExpr.Param[1].NumberLiteral.Int
			} else {
				minuteAllow = int64(funcExpr.Param[1].NumberLiteral.Float)
			}
		default:
		}
	}

	if minuteAllow > 15 {
		minuteAllow = 15
	} else if minuteAllow < 0 {
		minuteAllow = 0
	}

	minuteAllow *= int64(time.Minute)

	logTS, err := ng.GetKey(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	switch cont := logTS.Value.(type) {
	case int64:
		cont = detectTimezone(cont, time.Now().UnixNano(), minuteAllow)
		if err := ng.AddKey2PtWithVal(key, cont, ast.Int, runtime.KindPtDefault); err != nil {
			l.Debug(err)
			return nil
		}
	default:
		return fmt.Errorf("param value expect int64, got `%s`", reflect.TypeOf(cont))
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
