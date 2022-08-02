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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func JSONChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 4 {
		return fmt.Errorf("func %s expected 2 to 4 args", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if len(funcExpr.Param) == 3 {
		switch funcExpr.Param[2].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		default:
			return fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	if len(funcExpr.Param) == 4 {
		switch funcExpr.Param[3].(type) {
		case *parser.BoolLiteral:
		default:
			return fmt.Errorf("expect BoolLiteral, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	return nil
}

func JSON(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)

	var key, jpath, targetKey parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		jpath = v
	// TODO StringLiteral
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	targetKey = jpath
	if len(funcExpr.Param) == 3 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
			targetKey = v
		default:
			return fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	cont, err := ng.GetContentStr(key)
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
		switch v := funcExpr.Param[3].(type) {
		case *parser.BoolLiteral:
			trimSpace = v.Val
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

	if err := ng.SetContent(targetKey, v); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}

func GsonGet(s string, node interface{}) (interface{}, error) {
	var m interface{}

	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return "", err
	}

	return jsonGet(m, node)
}

func jsonGet(val interface{}, node interface{}) (interface{}, error) {
	switch t := node.(type) {
	case *parser.StringLiteral:
		return getByIdentifier(val, &parser.Identifier{Name: t.Val})
	case *parser.AttrExpr:
		return getByAttr(val, t)

	case *parser.Identifier:
		return getByIdentifier(val, t)

	case *parser.IndexExpr:
		child, err := getByIdentifier(val, t.Obj)
		if err != nil {
			return nil, err
		}
		return getByIndex(child, t, 0)
	default:
		return nil, fmt.Errorf("json unsupport get from %v", reflect.TypeOf(t))
	}
}

func getByAttr(val interface{}, i *parser.AttrExpr) (interface{}, error) {
	child, err := jsonGet(val, i.Obj)
	if err != nil {
		return nil, err
	}

	if i.Attr != nil {
		return jsonGet(child, i.Attr)
	}

	return child, nil
}

func getByIdentifier(val interface{}, i *parser.Identifier) (interface{}, error) {
	if i == nil {
		return val, nil
	}

	switch v := val.(type) {
	case map[string]interface{}:
		if child, ok := v[i.Name]; !ok {
			return nil, fmt.Errorf("%v not found", i.Name)
		} else {
			return child, nil
		}
	default:
		return nil, fmt.Errorf("%v unsupport identifier get", reflect.TypeOf(v))
	}
}

func getByIndex(val interface{}, i *parser.IndexExpr, dimension int) (interface{}, error) {
	switch v := val.(type) {
	case []interface{}:
		if dimension >= len(i.Index) {
			return nil, fmt.Errorf("dimension exceed")
		}

		index := int(i.Index[dimension])
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
