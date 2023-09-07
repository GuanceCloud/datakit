// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func AggCreateChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"bucket", "on_interval", "on_count",
		"keep_value", "const_tags", "category",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	arg := funcExpr.Param[0]
	switch arg.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `bucket` expect StringLiteral, got %s",
			arg.NodeType), arg.StartPos())
	}

	interval := time.Minute
	if arg := funcExpr.Param[1]; arg != nil {
		switch arg.NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			ts := arg.StringLiteral.Val
			if v, err := time.ParseDuration(ts); err != nil {
				return runtime.NewRunError(ctx, fmt.Sprintf("parse on_interval: %s", err.Error()),
					arg.StartPos())
			} else {
				interval = v
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param `on_interval` expect StringLiteral, got %s",
				arg.NodeType), arg.StartPos())
		}
	}

	count := 0
	if arg := funcExpr.Param[2]; arg != nil {
		switch arg.NodeType { //nolint:exhaustive
		case ast.TypeIntegerLiteral:
			count = int(arg.IntegerLiteral.Val)
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param `on_count` expect IntegerLiteral, got %s",
				arg.NodeType), arg.StartPos())
		}
	}

	if interval <= 0 && count <= 0 {
		return runtime.NewRunError(ctx,
			"param `on_interval` and `on_count` cannot be less than or equal to 0 at the same time", arg.StartPos())
	}

	if arg := funcExpr.Param[3]; arg != nil {
		switch arg.NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param `keep_value` expect BoolLiteral, got %s",
				arg.NodeType), arg.StartPos())
		}
	}

	return nil
}

func AggCreate(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	pt, err := getPoint(ctx.InData())
	if err != nil {
		return nil
	}

	buks := pt.GetAggBuckets()
	if buks == nil {
		return nil
	}

	ptCat := point.Metric
	if arg := funcExpr.Param[5]; arg != nil {
		if catName, _, err := runtime.RunStmt(ctx, arg); err != nil {
			return nil
		} else if catName != nil {
			if catN, ok := catName.(string); ok {
				ptCat = ptCategory(catN)
				if ptCat == point.UnknownCategory {
					return nil
				}
			} else {
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"type of parameter expected to be int64, got %s",
					reflect.TypeOf(catName)), arg.StartPos())
			}
		}
	}

	bukName := funcExpr.Param[0].StringLiteral.Val
	if _, ok := buks.GetBucket(ptCat, bukName); ok {
		return nil
	}

	interval := time.Minute
	count := 0
	keepValue := false

	if arg := funcExpr.Param[1]; arg != nil {
		interval, _ = time.ParseDuration(arg.StringLiteral.Val)
	}

	if arg := funcExpr.Param[2]; arg != nil {
		count = int(arg.IntegerLiteral.Val)
	}

	if arg := funcExpr.Param[3]; arg != nil {
		keepValue = arg.BoolLiteral.Val
	}

	constTags := map[string]string{}
	if arg := funcExpr.Param[4]; arg != nil {
		if v, _, err := runtime.RunStmt(ctx, arg); err == nil {
			if v, ok := v.(map[string]any); ok {
				for k, v := range v {
					if v, ok := v.(string); ok {
						constTags[k] = v
					}
				}
			}
		}
	}

	buks.CreateBucket(ptCat, bukName, interval, count, keepValue, constTags)

	return nil
}

func AggAddMetricChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"bucket", "new_field", "agg_fn",
		"agg_by", "agg_field", "category",
	}, 5); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	arg1 := funcExpr.Param[0]
	switch arg1.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `bucket` expect StringLiteral, got %s",
			arg1.NodeType), arg1.StartPos())
	}

	arg2 := funcExpr.Param[1]
	switch arg2.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `new_field` expect StringLiteral, got %s",
			arg2.NodeType), arg2.StartPos())
	}

	arg3 := funcExpr.Param[2]
	switch arg3.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `agg_fn` expect StringLiteral, got %s",
			arg3.NodeType), arg3.StartPos())
	}

	var tags []string
	arg4 := funcExpr.Param[3]
	switch arg4.NodeType { //nolint:exhaustive
	case ast.TypeListInitExpr:
		for _, v := range arg4.ListInitExpr.List {
			switch v.NodeType { //nolint:exhaustive
			case ast.TypeStringLiteral:
				tags = append(tags, v.StringLiteral.Val)
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf("agg_by elem expect StringLiteral, got %s",
					arg4.NodeType), arg4.StartPos())
			}
		}
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `agg_by` expect StringLiteral, got %s",
			arg4.NodeType), arg4.StartPos())
	}

	if len(tags) == 0 {
		return runtime.NewRunError(ctx, "size of param `agg_by` is 0", arg4.StartPos())
	}

	funcExpr.PrivateData = tags

	arg5 := funcExpr.Param[4]
	switch arg5.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("param `agg_field` expect StringLiteral, got %s",
			arg5.NodeType), arg5.StartPos())
	}

	return nil
}

func AggAddMetric(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	pt, err := getPoint(ctx.InData())
	if err != nil {
		return nil
	}

	buks := pt.GetAggBuckets()
	if buks == nil {
		return nil
	}

	bukName := funcExpr.Param[0].StringLiteral.Val
	newField := funcExpr.Param[1].StringLiteral.Val
	aggFn := funcExpr.Param[2].StringLiteral.Val

	aggBy, ok := funcExpr.PrivateData.([]string)
	if !ok {
		return nil
	}

	ptCat := point.Metric
	if arg := funcExpr.Param[5]; arg != nil {
		if catName, _, err := runtime.RunStmt(ctx, arg); err != nil {
			return nil
		} else if catName != nil {
			if catN, ok := catName.(string); ok {
				ptCat = ptCategory(catN)
				if ptCat == point.UnknownCategory {
					return nil
				}
			} else {
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"type of parameter expected to be int64, got %s",
					reflect.TypeOf(catName)), arg.StartPos())
			}
		}
	}

	byValue := []string{}
	for _, by := range aggBy {
		if v, err := ctx.GetKey(by); err != nil {
			return nil
		} else {
			if v, ok := v.Value.(string); ok {
				byValue = append(byValue, v)
			} else {
				return nil
			}
		}
	}

	if len(aggBy) != len(byValue) {
		return nil
	}

	aggField := funcExpr.Param[4].StringLiteral.Val

	fieldValue, err := ctx.GetKey(aggField)
	if err != nil {
		return nil
	}

	buk, ok := buks.GetBucket(ptCat, bukName)
	if !ok {
		return nil
	}

	buk.AddMetric(newField, aggFn, aggBy, byValue, fieldValue.Value)

	return nil
}
