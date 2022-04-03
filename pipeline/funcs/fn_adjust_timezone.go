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

func AdjustTimezoneChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}
	switch funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	return nil
}

func AdjustTimezone(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}
	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return fmt.Errorf("param key expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}
	cont, err := ng.GetContent(key)
	if err != nil {
		l.Debug(err)
		return nil
	}
	tn, err := detactTimezone(cont)
	if err != nil {
		return err
	}
	if err := ng.SetContent(key, tn); err != nil {
		l.Warn(err)
		return nil
	}
	return nil
}

const (
	timeHourNanosec  = int64(time.Hour)
	time45MinNanosec = int64(time.Minute * 45)
)

func detactTimezone(cont interface{}) (int64, error) {
	switch cont := cont.(type) {
	case int64:
		tn := time.Now().UnixNano()
		deltaTZ := (tn - cont - (tn%timeHourNanosec - cont%timeHourNanosec)) / timeHourNanosec
		if (deltaTZ > 24) || (deltaTZ < -24) {
			return 0, fmt.Errorf("delta time > 24h")
		}

		// 分钟进位
		deltaMin := tn%timeHourNanosec - cont%timeHourNanosec
		if deltaMin >= time45MinNanosec {
			deltaTZ += 1
		} else if deltaMin <= -time45MinNanosec {
			deltaTZ -= 1
		}

		return cont + deltaTZ*timeHourNanosec, nil
	default:
		return 0, fmt.Errorf("param value expect int64, got `%s`", reflect.TypeOf(cont))
	}
}
