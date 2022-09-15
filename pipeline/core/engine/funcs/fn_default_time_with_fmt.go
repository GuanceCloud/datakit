// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func DefaultTimeWithFmtChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
	default:
		return fmt.Errorf("param key expect AttrExpr or Identifier, got %s",
			funcExpr.Param[0].NodeType)
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return fmt.Errorf("param key expect StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}

	if len(funcExpr.Param) > 2 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType)
		}
	}

	return nil
}

func DefaultTimeWithFmt(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var err error
	var goTimeFmt string
	var tz string
	var t time.Time
	timezone := time.Local

	if len(funcExpr.Param) < 2 {
		return fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		goTimeFmt = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("param key expect StringLiteral, got %s",
			funcExpr.Param[1])
	}

	if len(funcExpr.Param) > 2 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			tz = funcExpr.Param[2].StringLiteral.Val
		default:
			return fmt.Errorf("param key expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType)
		}
	}

	timeStr, err := ng.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if tz != "" {
		timezone, err = time.LoadLocation(tz)
	}

	if err == nil {
		t, err = time.ParseInLocation(goTimeFmt, timeStr, timezone)
	}

	if err != nil {
		l.Debugf("time string: %s, time format: %s, timezone: %s, error msg: %s",
			timeStr, goTimeFmt, tz, err)
		return err
	} else {
		if err := ng.AddKey2PtWithVal(key, t.UnixNano(), ast.Int, runtime.KindPtDefault); err != nil {
			l.Debug(err)
			return nil
		}
		return nil
	}
}
