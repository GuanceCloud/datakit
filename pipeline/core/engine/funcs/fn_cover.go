// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func CoverChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	if funcExpr.Param[1].NodeType != ast.TypeListInitExpr {
		return fmt.Errorf("param range expects ListInitExpr, got %s", funcExpr.Param[1].NodeType)
	}

	set := funcExpr.Param[1].ListInitExpr

	if len(set.List) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", funcExpr.Param[1])
	}

	if set.List[0].NodeType != ast.TypeNumberLiteral ||
		set.List[1].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	}

	return nil
}

func Cover(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 2 {
		return fmt.Errorf("func %s expects 2 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	if funcExpr.Param[1].NodeType != ast.TypeListInitExpr {
		return nil
	}

	set := funcExpr.Param[1].ListInitExpr

	var start, end int

	if len(set.List) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if set.List[0].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else if set.List[0].NumberLiteral.IsInt {
		start = int(set.List[0].NumberLiteral.Int)
	}

	if set.List[1].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else if set.List[1].NumberLiteral.IsInt {
		end = int(set.List[1].NumberLiteral.Int)
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
		return fmt.Errorf("invalid cover range")
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

	if err := ctx.AddKey2PtWithVal(key, string(arrCont), ast.String,
		runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}
