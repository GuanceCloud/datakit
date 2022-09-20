// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func JSONChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 2 to 4 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType)
	}

	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
		default:
			return fmt.Errorf("expect AttrExpr or Identifier, got %s",
				funcExpr.Param[2].NodeType)
		}
	}

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return fmt.Errorf("expect BoolLiteral, got %s",
				funcExpr.Param[3].NodeType)
		}
	}

	return nil
}

func JSON(ctx *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var jpath *ast.Node

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
		jpath = funcExpr.Param[1]
	// TODO StringLiteral
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType)
	}

	targetKey, _ := getKeyName(jpath)

	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
			targetKey, _ = getKeyName(funcExpr.Param[2])
		default:
			return fmt.Errorf("expect AttrExpr or Identifier, got %s",
				funcExpr.Param[2].NodeType)
		}
	}

	cont, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}

	v, err := GsonGet(cont, jpath)
	if err != nil {
		l.Debug(err)
		return nil
	}

	var trimSpace bool
	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			trimSpace = funcExpr.Param[3].BoolLiteral.Val
		default:
			return fmt.Errorf("expect BoolLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	} else {
		trimSpace = true
	}

	if vStr, ok := v.(string); ok && trimSpace {
		v = strings.TrimSpace(vStr)
	}

	var dtype ast.DType
	switch v.(type) {
	case bool:
		dtype = ast.Bool
	case float64:
		dtype = ast.Float
	case string:
		dtype = ast.String
	case []any:
		dtype = ast.List
	case map[string]any:
		dtype = ast.Map
	default:
		return nil
	}
	if err := ctx.AddKey2PtWithVal(targetKey, v, dtype, runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}

func GsonGet(s string, node *ast.Node) (any, error) {
	var m any

	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return "", err
	}

	return jsonGet(m, node)
}

func jsonGet(val any, node *ast.Node) (any, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		return getByIdentifier(val, &ast.Identifier{
			Name: node.StringLiteral.Val,
		})
	case ast.TypeAttrExpr:
		return getByAttr(val, node.AttrExpr)

	case ast.TypeIdentifier:
		return getByIdentifier(val, node.Identifier)

	case ast.TypeIndexExpr:
		child, err := getByIdentifier(val, node.IndexExpr.Obj)
		if err != nil {
			return nil, err
		}
		return getByIndex(child, node.IndexExpr, 0)
	default:
		return nil, fmt.Errorf("json unsupport get from %s", node.NodeType)
	}
}

func getByAttr(val any, i *ast.AttrExpr) (any, error) {
	child, err := jsonGet(val, i.Obj)
	if err != nil {
		return nil, err
	}

	if i.Attr != nil {
		return jsonGet(child, i.Attr)
	}

	return child, nil
}

func getByIdentifier(val any, i *ast.Identifier) (any, error) {
	if i == nil {
		return val, nil
	}

	switch v := val.(type) {
	case map[string]any:
		if child, ok := v[i.Name]; !ok {
			return nil, fmt.Errorf("%v not found", i.Name)
		} else {
			return child, nil
		}
	default:
		return nil, fmt.Errorf("%v unsupport identifier get", reflect.TypeOf(v))
	}
}

func getByIndex(val any, i *ast.IndexExpr, dimension int) (any, error) {
	switch v := val.(type) {
	case []any:
		if dimension >= len(i.Index) {
			return nil, fmt.Errorf("dimension exceed")
		}

		if i.Index[dimension].NodeType != ast.TypeNumberLiteral {
			return nil, fmt.Errorf("index value is not int")
		}
		var index int
		if i.Index[dimension].NumberLiteral.IsInt {
			index = int(i.Index[dimension].NumberLiteral.Int)
		} else {
			index = int(i.Index[dimension].NumberLiteral.Float)
		}
		if index < 0 {
			index = len(v) + index
		}

		if index < 0 || index >= len(v) {
			return nil, fmt.Errorf("index out of range")
		}

		child := v[index]
		if dimension == len(i.Index)-1 {
			return child, nil
		} else {
			return getByIndex(child, i, dimension+1)
		}
	default:
		return nil, fmt.Errorf("%v unsupport index get", reflect.TypeOf(v))
	}
}
