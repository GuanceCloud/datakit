// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func CoverChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if funcExpr.Param[1].NodeType != ast.TypeListLiteral {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param range expects ListInitExpr, got %s", funcExpr.Param[1].NodeType),
			funcExpr.Param[1].StartPos())
	}

	set := funcExpr.Param[1].ListLiteral()

	if len(set.List) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param between range value `%v' is not expected", funcExpr.Param[1]),
			funcExpr.Param[1].StartPos())
	}

	if (set.List[0].NodeType != ast.TypeFloatLiteral && set.List[0].NodeType != ast.TypeIntegerLiteral) ||
		(set.List[1].NodeType != ast.TypeFloatLiteral && set.List[1].NodeType != ast.TypeIntegerLiteral) {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value `%v' is not expected", set), funcExpr.Param[1].StartPos())
	}

	return nil
}

func Cover(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expects 2 args", funcExpr.Name), funcExpr.NamePos)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	if funcExpr.Param[1].NodeType != ast.TypeListLiteral {
		return nil
	}

	set := funcExpr.Param[1].ListLiteral()

	var start, end int

	if len(set.List) != 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"param between range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	switch set.List[0].NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral:
		start = int(set.List[0].IntegerLiteral().Val)
	case ast.TypeFloatLiteral:
		start = int(set.List[0].FloatLiteral().Val)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	switch set.List[1].NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral:
		end = int(set.List[1].IntegerLiteral().Val)
	case ast.TypeFloatLiteral:
		end = int(set.List[1].FloatLiteral().Val)
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"range value `%v' is not expected", set),
			funcExpr.Param[1].StartPos())
	}

	cont1, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var cont string

	switch v := cont1.Value.(type) {
	case string:
		cont = v
	default:
		return nil
	}

	if end > utf8.RuneCountInString(cont) {
		end = utf8.RuneCountInString(cont)
	}

	// end less than 0  become greater than 0
	if end < 0 {
		end += utf8.RuneCountInString(cont) + 1
	}
	// start less than 0  become greater than 0
	if start <= 0 {
		start += utf8.RuneCountInString(cont) + 1
	}

	// unreasonable subscript
	if start > end {
		l.Debug("invalid cover range")
		return nil
		// return runtime.NewRunError(ctx, "invalid cover range", funcExpr.Param[1].StartPos())
	}

	arrCont := []rune(cont)

	for i := 0; i < len(arrCont); i++ {
		if i+1 >= start && i < end {
			if unicode.Is(unicode.Han, arrCont[i]) {
				arrCont[i] = rune('＊')
			} else {
				arrCont[i] = rune('*')
			}
		}
	}

	_ = addKey2PtWithVal(ctx.InData(), key, string(arrCont),
		ast.String, ptinput.KindPtDefault)

	return nil
}
