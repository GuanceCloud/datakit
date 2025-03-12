// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"strconv"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/spf13/cast"
)

func ParseDateChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "yy", "MM", "dd", "hh", "mm", "ss", "ms", "us", "ns", "zone",
	}, 0); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}
	return nil
}

func ParseDate(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	now := time.Now()
	var key string

	var zone *time.Location

	var yy int
	var MM time.Month
	var dd, hh, mm, ss, ms, us, ns int

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	// year
	if x := getArgStrConv2Int(ctx, funcExpr.Param[1]); true {
		yy, err = fixYear(now, x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
	}

	// month
	if x := getArgStr(ctx, funcExpr.Param[2]); true {
		MM, err = fixMonth(now, x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
		}
	}

	// day
	if x := getArgStrConv2Int(ctx, funcExpr.Param[3]); true {
		dd, err = fixDay(now, x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[3].StartPos())
		}
	}

	// hour
	if x := getArgStrConv2Int(ctx, funcExpr.Param[4]); true {
		hh, err = fixHour(now, x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[4].StartPos())
		}
	}
	// mm
	if x := getArgStrConv2Int(ctx, funcExpr.Param[5]); true {
		mm, err = fixMinute(now, x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[5].StartPos())
		}
	}

	// ss
	if x := getArgStrConv2Int(ctx, funcExpr.Param[6]); true {
		ss, err = fixSecond(x)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[6].StartPos())
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
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[10].StartPos())
			}
		}
	}

	t := time.Date(yy, MM, dd, hh, mm, ss, ms*1000*1000+us*1000+ns, zone)

	_ = addKey2PtWithVal(ctx.InData(), key, t.UnixNano(), ast.Int, ptinput.KindPtDefault)

	return nil
}

func getArgStr(ctx *runtime.Task, node *ast.Node) string {
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

func getArgStrConv2Int(ctx *runtime.Task, node *ast.Node) int64 {
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
