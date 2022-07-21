// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

const defaultMinuteDelta = int64(2)

func AdjustTimezoneChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)

	if len(funcExpr.Param) < 1 {
		return fmt.Errorf("func `%s' expected 1 or 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if len(funcExpr.Param) == 2 {
		switch funcExpr.Param[1].(type) {
		case *parser.NumberLiteral:
		default:
			return fmt.Errorf("param key expect NumberLiteral, got '%s' ",
				reflect.TypeOf(funcExpr.Param[1].String()))
		}
	}
	return nil
}

func AdjustTimezone(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	// 默认允许 +2 分钟误差
	minuteAllow := defaultMinuteDelta
	if len(funcExpr.Param) == 2 {
		switch v := funcExpr.Param[1].(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				minuteAllow = v.Int
			} else {
				minuteAllow = int64(v.Float)
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

	logTS, err := ng.GetContent(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	switch cont := logTS.(type) {
	case int64:
		cont = detectTimezone(cont, time.Now().UnixNano(), minuteAllow)
		if err := ng.SetContent(key, cont); err != nil {
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
