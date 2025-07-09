// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ast

import (
	"github.com/spf13/cast"
)

type DType uint

func (t DType) String() string {
	switch t {
	case Invalid:
		return "invalid"
	case Void:
		return "void"
	case Nil:
		return "nil"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Float:
		return "float"
	case String:
		return "str"
	case List:
		return "list"
	case Map:
		return "map"
	}
	return ""
}

const (
	Invalid DType = iota

	Void // no return value

	Nil // nil

	Bool
	Int    // int64
	Float  // float64
	String // string

	List // []any

	// map[string]any (load json to string map or any array).
	Map
	//// or map[any]any (default).
	// Map.
)

func AllTyp() []DType {
	return []DType{Bool, Int, Float, String, List, Map}
}

func DectDataType(val any) (any, DType) {
	if val == nil {
		return nil, Nil
	}

	switch val.(type) {
	case string:
		return val, String
	case int, int16, int32, int8,
		uint, uint16, uint32, uint64, uint8:
		return cast.ToInt64(val), Int
	case int64:
		return val, Int
	case float32:
		return cast.ToFloat64(val), Float
	case float64:
		return val, Float
	case bool:
		return val, Bool
	case map[string]any:
		return val, Map
	// case map[any]any:
	// 	return Map
	case []any:
		return val, List
	default:
		return nil, Invalid
	}
}
