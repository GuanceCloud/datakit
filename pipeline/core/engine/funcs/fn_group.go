// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func GroupChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	var start, end float64

	if len(funcExpr.Param) == 4 {
		if _, err := getKeyName(funcExpr.Param[3]); err != nil {
			return err
		}
	}

	var set []*ast.Node
	if funcExpr.Param[1].NodeType == ast.TypeListInitExpr {
		set = funcExpr.Param[1].ListInitExpr.List
	}

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if set[0].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if set[0].NumberLiteral.IsInt {
			start = float64(set[0].NumberLiteral.Int)
		} else {
			start = set[0].NumberLiteral.Float
		}
	}

	if set[1].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if set[1].NumberLiteral.IsInt {
			end = float64(set[1].NumberLiteral.Int)
		} else {
			end = set[1].NumberLiteral.Float
		}

		if start > end {
			return fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}
	return nil
}

func Group(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
	}

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ctx.GetKey(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	var start, end float64

	var set []*ast.Node
	if funcExpr.Param[1].NodeType == ast.TypeListInitExpr {
		set = funcExpr.Param[1].ListInitExpr.List
	}

	if len(set) != 2 {
		return fmt.Errorf("param between range value `%v' is not expected", set)
	}

	if set[0].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if set[0].NumberLiteral.IsInt {
			start = float64(set[0].NumberLiteral.Int)
		} else {
			start = set[0].NumberLiteral.Float
		}
	}

	if set[1].NodeType != ast.TypeNumberLiteral {
		return fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if set[1].NumberLiteral.IsInt {
			end = float64(set[1].NumberLiteral.Int)
		} else {
			end = set[1].NumberLiteral.Float
		}

		if start > end {
			return fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}

	if GroupHandle(cont.Value, start, end) {
		value := funcExpr.Param[2]

		var val any
		var dtype ast.DType

		switch value.NodeType { //nolint:exhaustive
		case ast.TypeNumberLiteral:
			if value.NumberLiteral.IsInt {
				val = value.NumberLiteral.Int
				dtype = ast.Int
			} else {
				val = value.NumberLiteral.Float
				dtype = ast.Float
			}
		case ast.TypeStringLiteral:
			val = value.StringLiteral.Val
			dtype = ast.String
		case ast.TypeBoolLiteral:
			val = value.BoolLiteral.Val
			dtype = ast.String
		default:
			l.Debugf("unknown group elements: %s", value.NodeType)
			return fmt.Errorf("unsupported group type")
		}

		if len(funcExpr.Param) == 4 {
			if k, err := getKeyName(funcExpr.Param[3]); err == nil {
				key = k
			} else {
				return err
			}
		}
		_ = ctx.AddKey2PtWithVal(key, val, dtype, runtime.KindPtDefault)
	}

	return nil
}
