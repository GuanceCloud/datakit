// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"strconv"
	"time"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func ParseDateChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if err := reIndexFuncArgs(funcExpr, []string{
		"key", "yy", "MM", "dd", "hh", "mm", "ss", "ms", "us", "ns", "zone",
	}, 0); err != nil {
		return err
	}
	return nil
}

func ParseDate(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	now := time.Now()
	var key string

	var zone *time.Location

	var yy int
	var MM time.Month
	var dd, hh, mm, ss, ms, us, ns int

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	// year
	if x := getArgStrConv2Int(ctx, funcExpr.Param[1]); true {
		yy, err = fixYear(now, x)
		if err != nil {
			return err
		}
	}

	// month
	if x := getArgStr(ctx, funcExpr.Param[2]); true {
		MM, err = fixMonth(now, x)
		if err != nil {
			return err
		}
	}

	// day
	if x := getArgStrConv2Int(ctx, funcExpr.Param[3]); true {
		dd, err = fixDay(now, x)
		if err != nil {
			return err
		}
	}

	// hour
	if x := getArgStrConv2Int(ctx, funcExpr.Param[4]); true {
		hh, err = fixHour(now, x)
		if err != nil {
			return err
		}
	}
	// mm
	if x := getArgStrConv2Int(ctx, funcExpr.Param[5]); true {
		mm, err = fixMinute(now, x)
		if err != nil {
			return err
		}
	}

	// ss
	if x := getArgStrConv2Int(ctx, funcExpr.Param[6]); true {
		ss, err = fixSecond(x)
		if err != nil {
			return err
		}
	}
	// ms
	if x := getArgStrConv2Int(ctx, funcExpr.Param[7]); true {
		if x == DefaultInt {
			ms = 0
		} else {
			ms = int(x)
		}
	}

	// us
	if x := getArgStrConv2Int(ctx, funcExpr.Param[8]); true {
		if x == DefaultInt {
			us = 0
		} else {
			us = int(x)
		}
	}

	// ns
	if x := getArgStrConv2Int(ctx, funcExpr.Param[9]); true {
		if x == DefaultInt {
			ns = 0
		} else {
			ns = int(x)
		}
	}

	if x := getArgStr(ctx, funcExpr.Param[10]); true {
		if x == "" {
			zone = time.UTC
		} else {
			zone, err = tz(x)
			if err != nil {
				return err
			}
		}
	}

	t := time.Date(yy, MM, dd, hh, mm, ss, ms*1000*1000+us*1000+ns, zone)

	_ = ctx.AddKey2PtWithVal(key, t.UnixNano(), ast.Int, runtime.KindPtDefault)

	return nil
}

func getArgStr(ctx *runtime.Context, node *ast.Node) string {
	if node == nil {
		return ""
	}

	if v, dtype, err := runtime.RunStmt(ctx, node); err == nil {
		if dtype == ast.String {
			return cast.ToString(v)
		}
	}
	return ""
}

func getArgStrConv2Int(ctx *runtime.Context, node *ast.Node) int64 {
	str := getArgStr(ctx, node)
	if str == "" {
		return DefaultInt
	}

	v, err := strconv.ParseInt(str, 10, 64) //nolint: gomnd
	if err != nil {
		return DefaultInt
	}
	return v
}
