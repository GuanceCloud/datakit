// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-json"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func JSONChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"input", "json_path", "newkey",
		"trim_space", "delete_after_extract",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	lastIdxExpr := false
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeIndexExpr:
		var err error
		lastIdxExpr, err = lastIsIndex(funcExpr.Param[1])
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr, IndexExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	if funcExpr.Param[2] != nil {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	if funcExpr.Param[3] != nil {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.Param[3].StartPos())
		}
	}

	if funcExpr.Param[4] != nil {
		switch funcExpr.Param[4].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			if funcExpr.Param[4].BoolLiteral().Val == lastIdxExpr {
				return runtime.NewRunError(ctx, "does not support deleting elements in the list",
					funcExpr.Param[4].StartPos())
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.Param[4].StartPos())
		}
	}

	return nil
}

func JSON(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var jpath *ast.Node

	srcKey, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeIndexExpr:
		jpath = funcExpr.Param[1]
	// TODO StringLiteral
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	targetKey, _ := getKeyName(jpath)

	if funcExpr.Param[2] != nil {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
			targetKey, _ = getKeyName(funcExpr.Param[2])
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
				funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
		}
	}

	cont, err := ctx.GetKeyConv2Str(srcKey)
	if err != nil {
		l.Debug(err)
		return nil
	}

	deleteAfterExtract := false
	if funcExpr.Param[4] != nil {
		switch funcExpr.Param[4].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			deleteAfterExtract = funcExpr.Param[4].BoolLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.Param[4].StartPos())
		}
	}

	v, dstS, err := GsonGet(cont, jpath, deleteAfterExtract)
	if err != nil {
		l.Debug(err)
		return nil
	}

	trimSpace := true
	if funcExpr.Param[3] != nil {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeBoolLiteral:
			trimSpace = funcExpr.Param[3].BoolLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.Param[3].StartPos())
		}
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
	if ok := addKey2PtWithVal(ctx.InData(), targetKey, v, dtype, ptinput.KindPtDefault); !ok {
		return nil
	}

	if deleteAfterExtract {
		_ = addKey2PtWithVal(ctx.InData(), srcKey, dstS, ast.String, ptinput.KindPtDefault)
	}

	return nil
}

func GsonGet(s string, node *ast.Node, deleteAfter bool) (any, string, error) {
	var m any

	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return "", "", err
	}

	val, err := jsonGet(m, node, deleteAfter)
	if err != nil {
		return "", "", err
	}

	dst := s
	if deleteAfter {
		dstB, err := json.Marshal(m)
		if err != nil {
			return "", "", err
		}
		dst = string(dstB)
	}
	return val, dst, nil
}

func jsonGet(val any, node *ast.Node, deleteAfter bool) (any, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		return getByIdentifier(val, &ast.Identifier{
			Name: node.StringLiteral().Val,
		}, deleteAfter)
	case ast.TypeAttrExpr:
		return getByAttr(val, node.AttrExpr(), deleteAfter)

	case ast.TypeIdentifier:
		return getByIdentifier(val, node.Identifier(), deleteAfter)

	case ast.TypeIndexExpr:
		child, err := getByIdentifier(val, node.IndexExpr().Obj, false)
		if err != nil {
			return nil, err
		}
		return getByIndex(child, node.IndexExpr(), 0, deleteAfter)
	default:
		return nil, fmt.Errorf("json unsupport get from %s", node.NodeType)
	}
}

func getByAttr(val any, i *ast.AttrExpr, deleteAfter bool) (any, error) {
	if i.Attr != nil {
		child, err := jsonGet(val, i.Obj, false)
		if err != nil {
			return nil, err
		}
		return jsonGet(child, i.Attr, deleteAfter)
	} else {
		child, err := jsonGet(val, i.Obj, deleteAfter)
		if err != nil {
			return nil, err
		}
		return child, nil
	}
}

func getByIdentifier(val any, i *ast.Identifier, deleteAfter bool) (any, error) {
	if i == nil {
		return val, nil
	}

	switch v := val.(type) {
	case map[string]any:
		if child, ok := v[i.Name]; !ok {
			return nil, fmt.Errorf("%v not found", i.Name)
		} else {
			if deleteAfter {
				delete(v, i.Name)
			}
			return child, nil
		}
	default:
		return nil, fmt.Errorf("%v unsupport identifier get", reflect.TypeOf(v))
	}
}

func getByIndex(val any, i *ast.IndexExpr, dimension int, deleteAfter bool) (any, error) {
	switch v := val.(type) {
	case []any:
		if dimension >= len(i.Index) {
			return nil, fmt.Errorf("dimension exceed")
		}

		var index int

		switch i.Index[dimension].NodeType { //nolint:exhaustive
		case ast.TypeIntegerLiteral:
			index = int(i.Index[dimension].IntegerLiteral().Val)
		case ast.TypeFloatLiteral:
			index = int(i.Index[dimension].FloatLiteral().Val)

		default:
			return nil, fmt.Errorf("index value is not int")
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
			return getByIndex(child, i, dimension+1, deleteAfter)
		}
	default:
		return nil, fmt.Errorf("%v unsupport index get", reflect.TypeOf(v))
	}
}

func lastIsIndex(expr *ast.Node) (bool, error) {
	switch expr.NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr:
		return lastIsIndex(expr.AttrExpr().Attr)
	case ast.TypeIdentifier:
		return false, nil
	case ast.TypeIndexExpr:
		return true, nil
	default:
		return false, fmt.Errorf("expect AttrExpr, IndexExpr or Identifier, got %s",
			expr.NodeType)
	}
}
