// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func GroupChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 3 or 4 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	var start, end float64

	if len(funcExpr.Param) == 4 {
		if _, err := getKeyName(funcExpr.Param[3]); err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[3].StartPos())
		}
	}

	var set []*ast.Node
	if funcExpr.Param[1].NodeType == ast.TypeListLiteral {
		set = funcExpr.Param[1].ListLiteral().List
	}

	if len(set) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param between range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	switch set[0].NodeType { //nolint:exhaustive
	case ast.TypeFloatLiteral:
		start = set[0].FloatLiteral().Val
	case ast.TypeIntegerLiteral:
		start = float64(set[0].IntegerLiteral().Val)

	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value `%v' is not expected", set), set[0].StartPos())
	}

	switch set[1].NodeType { //nolint:exhaustive
	case ast.TypeFloatLiteral:
		end = set[1].FloatLiteral().Val
	case ast.TypeIntegerLiteral:
		end = float64(set[1].IntegerLiteral().Val)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value `%v' is not expected", set), set[1].StartPos())
	}

	if start > end {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value start %v must le end %v", start, end), funcExpr.NamePos)
	}

	return nil
}

func Group(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func `%s' expected 3 or 4 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var start, end float64

	var set []*ast.Node
	if funcExpr.Param[1].NodeType == ast.TypeListLiteral {
		set = funcExpr.Param[1].ListLiteral().List
	}

	if len(set) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param between range value `%v' is not expected", set), funcExpr.Param[1].StartPos())
	}

	switch set[0].NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral:
		start = float64(set[0].IntegerLiteral().Val)
	case ast.TypeFloatLiteral:
		start = set[0].FloatLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	switch set[1].NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral:
		end = float64(set[1].IntegerLiteral().Val)
	case ast.TypeFloatLiteral:
		end = set[1].FloatLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	if start > end {
		return runtime.NewRunError(ctx, fmt.Sprintf("range value start %v must le end %v", start, end),
			funcExpr.Param[1].StartPos())
	}

	if GroupHandle(cont.Value, start, end) {
		value := funcExpr.Param[2]

		var val any
		var dtype ast.DType

		switch value.NodeType { //nolint:exhaustive
		case ast.TypeFloatLiteral:
			val = value.FloatLiteral().Val
			dtype = ast.Float
		case ast.TypeIntegerLiteral:
			val = value.IntegerLiteral().Val
			dtype = ast.Int
		case ast.TypeStringLiteral:
			val = value.StringLiteral().Val
			dtype = ast.String
		case ast.TypeBoolLiteral:
			val = value.BoolLiteral().Val
			dtype = ast.String
		default:
			l.Debugf("unknown group elements: %s", value.NodeType)
			return runtime.NewRunError(ctx, "unsupported group type",
				funcExpr.NamePos)
		}

		if len(funcExpr.Param) == 4 {
			if k, err := getKeyName(funcExpr.Param[3]); err == nil {
				key = k
			} else {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[4].StartPos())
			}
		}
		_ = addKey2PtWithVal(ctx.InData(), key, val, dtype, ptinput.KindPtDefault)
	}

	return nil
}
