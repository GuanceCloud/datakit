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
	"github.com/tidwall/gjson"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

type compiledGJSONPath struct {
	parts []string
	ok    bool
}

type compiledJSONCall struct {
	srcKey             string
	targetKey          string
	trimSpace          bool
	deleteAfterExtract bool
	gjsonPath          *compiledGJSONPath
}

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

	funcExpr.PrivateData = compileJSONCallData(funcExpr)

	return nil
}

func JSON(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var jpath *ast.Node
	var compiled *compiledJSONCall

	srcKey := ""
	targetKey := ""
	deleteAfterExtract := false
	trimSpace := true

	if c := getCompiledJSONCall(funcExpr); c != nil {
		compiled = c
		srcKey = c.srcKey
		targetKey = c.targetKey
		deleteAfterExtract = c.deleteAfterExtract
		trimSpace = c.trimSpace
	} else {
		var err error
		srcKey, err = getKeyName(funcExpr.Param[0])
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
		}
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeIndexExpr:
		jpath = funcExpr.Param[1]
	// TODO StringLiteral
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	if compiled == nil {
		targetKey, _ = getKeyName(jpath)

		if funcExpr.Param[2] != nil {
			switch funcExpr.Param[2].NodeType { //nolint:exhaustive
			case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
				targetKey, _ = getKeyName(funcExpr.Param[2])
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf("expect AttrExpr or Identifier, got %s",
					funcExpr.Param[2].NodeType), funcExpr.Param[2].StartPos())
			}
		}
	}

	cont, err := ctx.GetKeyConv2Str(srcKey)
	if err != nil {
		l.Debug(err)
		return nil
	}

	if compiled == nil {
		if funcExpr.Param[4] != nil {
			switch funcExpr.Param[4].NodeType { //nolint:exhaustive
			case ast.TypeBoolLiteral:
				deleteAfterExtract = funcExpr.Param[4].BoolLiteral().Val
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
					funcExpr.Param[3].NodeType), funcExpr.Param[4].StartPos())
			}
		}

		if funcExpr.Param[3] != nil {
			switch funcExpr.Param[3].NodeType { //nolint:exhaustive
			case ast.TypeBoolLiteral:
				trimSpace = funcExpr.Param[3].BoolLiteral().Val
			default:
				return runtime.NewRunError(ctx, fmt.Sprintf("expect BoolLiteral, got %s",
					funcExpr.Param[3].NodeType), funcExpr.Param[3].StartPos())
			}
		}
	}

	var (
		v     any
		dtype ast.DType
		dstS  string
	)

	if deleteAfterExtract {
		var err error
		v, dstS, err = GsonGet(cont, jpath, true)
		if err != nil {
			l.Debug(err)
			return nil
		}

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
	} else {
		var err error
		v, dtype, err = GJSONGet(cont, jpath, getCompiledGJSONPath(compiled))
		if err != nil {
			l.Debug(err)
			return nil
		}
		if dtype == ast.Nil || dtype == ast.Invalid {
			return nil
		}
	}

	if vStr, ok := v.(string); ok && trimSpace {
		v = strings.TrimSpace(vStr)
	}

	if ok := addKey2PtWithVal(ctx.InData(), targetKey, v, dtype, ptinput.KindPtDefault); !ok {
		return nil
	}

	if deleteAfterExtract {
		_ = addKey2PtWithVal(ctx.InData(), srcKey, dstS, ast.String, ptinput.KindPtDefault)
	}

	return nil
}

func compileJSONCallData(funcExpr *ast.CallExpr) *compiledJSONCall {
	srcKey, _ := getKeyName(funcExpr.Param[0])
	targetKey, _ := getKeyName(funcExpr.Param[1])
	if funcExpr.Param[2] != nil {
		targetKey, _ = getKeyName(funcExpr.Param[2])
	}

	trimSpace := true
	if funcExpr.Param[3] != nil {
		trimSpace = funcExpr.Param[3].BoolLiteral().Val
	}

	deleteAfterExtract := false
	if funcExpr.Param[4] != nil {
		deleteAfterExtract = funcExpr.Param[4].BoolLiteral().Val
	}

	return &compiledJSONCall{
		srcKey:             srcKey,
		targetKey:          targetKey,
		trimSpace:          trimSpace,
		deleteAfterExtract: deleteAfterExtract,
		gjsonPath:          compileGJSONPathData(funcExpr.Param[1]),
	}
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

func GJSONGet(s string, node *ast.Node, compiled *compiledGJSONPath) (any, ast.DType, error) {
	if !gjson.Valid(s) {
		return nil, ast.Invalid, fmt.Errorf("invalid json")
	}

	if compiled != nil && compiled.ok {
		res, err := gjsonGetByPathParts(gjson.Parse(s), compiled.parts)
		if err != nil {
			return nil, ast.Invalid, err
		}
		return gjsonResultValue(res)
	}

	res, err := gjsonGet(gjson.Parse(s), node)
	if err != nil {
		return nil, ast.Invalid, err
	}

	return gjsonResultValue(res)
}

func getCompiledJSONCall(funcExpr *ast.CallExpr) *compiledJSONCall {
	if funcExpr == nil {
		return nil
	}
	if compiled, ok := funcExpr.PrivateData.(*compiledJSONCall); ok {
		return compiled
	}
	return nil
}

func getCompiledGJSONPath(compiled *compiledJSONCall) *compiledGJSONPath {
	if compiled == nil {
		return nil
	}
	return compiled.gjsonPath
}

func compileGJSONPathData(node *ast.Node) *compiledGJSONPath {
	parts, ok := compileGJSONPathParts(node)
	return &compiledGJSONPath{
		parts: parts,
		ok:    ok,
	}
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

func gjsonGet(res gjson.Result, node *ast.Node) (gjson.Result, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeStringLiteral:
		return gjsonGetByIdentifier(res, node.StringLiteral().Val)
	case ast.TypeAttrExpr:
		return gjsonGetByAttr(res, node.AttrExpr())
	case ast.TypeIdentifier:
		return gjsonGetByIdentifier(res, node.Identifier().Name)
	case ast.TypeIndexExpr:
		child, err := gjsonGetByIdentifierNode(res, node.IndexExpr().Obj)
		if err != nil {
			return gjson.Result{}, err
		}
		return gjsonGetByIndex(child, node.IndexExpr(), 0)
	default:
		return gjson.Result{}, fmt.Errorf("json unsupport get from %s", node.NodeType)
	}
}

func compileGJSONPathParts(node *ast.Node) ([]string, bool) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		return compileGJSONPathPart(node.Identifier().Name)
	case ast.TypeStringLiteral:
		return compileGJSONPathPart(node.StringLiteral().Val)
	case ast.TypeAttrExpr:
		return compileGJSONAttrPathParts(node.AttrExpr())
	case ast.TypeIndexExpr:
		// gjson numeric path segments also match object keys like "0".
		// Keep IndexExpr on the strict walker so indexes only apply to arrays.
		return nil, false
	default:
		return nil, false
	}
}

func compileGJSONPathPart(part string) ([]string, bool) {
	if isGJSONArrayIndexPathPart(part) {
		return nil, false
	}
	return []string{part}, true
}

func compileGJSONAttrPathParts(expr *ast.AttrExpr) ([]string, bool) {
	if expr == nil {
		return nil, false
	}
	if expr.Attr == nil {
		return compileGJSONPathParts(expr.Obj)
	}

	var parts []string
	if expr.Obj != nil {
		objParts, ok := compileGJSONPathParts(expr.Obj)
		if !ok {
			return nil, false
		}
		parts = append(parts, objParts...)
	}

	attrParts, ok := compileGJSONPathParts(expr.Attr)
	if !ok {
		return nil, false
	}
	return append(parts, attrParts...), true
}

func isGJSONArrayIndexPathPart(part string) bool {
	if part == "" {
		return false
	}
	for _, ch := range part {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
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

func gjsonGetByAttr(res gjson.Result, i *ast.AttrExpr) (gjson.Result, error) {
	if i.Attr != nil {
		child, err := gjsonGet(res, i.Obj)
		if err != nil {
			return gjson.Result{}, err
		}
		return gjsonGet(child, i.Attr)
	}

	return gjsonGet(res, i.Obj)
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

func gjsonGetByIdentifierNode(res gjson.Result, i *ast.Identifier) (gjson.Result, error) {
	if i == nil {
		return res, nil
	}

	return gjsonGetByIdentifier(res, i.Name)
}

func gjsonGetByIdentifier(res gjson.Result, key string) (gjson.Result, error) {
	if !res.IsObject() {
		return gjson.Result{}, fmt.Errorf("%s unsupport identifier get", res.Type)
	}

	var (
		child gjson.Result
		found bool
	)
	res.ForEach(func(k, v gjson.Result) bool {
		if k.String() == key {
			child = v
			found = true
		}
		return true
	})
	if !found {
		return gjson.Result{}, fmt.Errorf("%v not found", key)
	}

	return child, nil
}

func gjsonGetByPathParts(res gjson.Result, parts []string) (gjson.Result, error) {
	for _, part := range parts {
		var err error
		res, err = gjsonGetByIdentifier(res, part)
		if err != nil {
			return gjson.Result{}, err
		}
	}
	return res, nil
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

func gjsonGetByIndex(res gjson.Result, i *ast.IndexExpr, dimension int) (gjson.Result, error) {
	if !res.IsArray() {
		return gjson.Result{}, fmt.Errorf("%s unsupport index get", res.Type)
	}

	if dimension >= len(i.Index) {
		return gjson.Result{}, fmt.Errorf("dimension exceed")
	}

	index, err := jsonIndex(i.Index[dimension])
	if err != nil {
		return gjson.Result{}, err
	}

	arr := res.Array()
	if index < 0 {
		index = len(arr) + index
	}

	if index < 0 || index >= len(arr) {
		return gjson.Result{}, fmt.Errorf("index out of range")
	}

	child := arr[index]
	if dimension == len(i.Index)-1 {
		return child, nil
	}

	return gjsonGetByIndex(child, i, dimension+1)
}

func jsonIndex(node *ast.Node) (int, error) {
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeIntegerLiteral:
		return int(node.IntegerLiteral().Val), nil
	case ast.TypeFloatLiteral:
		return int(node.FloatLiteral().Val), nil
	default:
		return 0, fmt.Errorf("index value is not int")
	}
}

func gjsonResultValue(res gjson.Result) (any, ast.DType, error) {
	switch res.Type {
	case gjson.Number:
		return res.Float(), ast.Float, nil
	case gjson.True, gjson.False:
		return res.Bool(), ast.Bool, nil
	case gjson.String:
		return res.String(), ast.String, nil
	case gjson.JSON:
		if res.IsObject() {
			return res.Value(), ast.Map, nil
		}
		if res.IsArray() {
			return res.Value(), ast.List, nil
		}
	case gjson.Null:
		return nil, ast.Nil, nil
	}

	return nil, ast.Invalid, fmt.Errorf("unsupported json result type %s", res.Type)
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
