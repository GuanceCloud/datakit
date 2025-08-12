// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"reflect"

	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
)

var (
	// EnableMixedArrayField and EnableDictField used to allow mix-typed array and dict/map
	// value in point field.
	//
	// Currently, GuanceDB backend do NOT support mix-typed array and dict.
	EnableMixedArrayField = false
	EnableDictField       = false
)

const (
	ArrayFieldType = "type.googleapis.com/point.Array"
	DictFieldType  = "type.googleapis.com/point.Map"
)

// MustAnyRaw get underlying wrapped value, and panic if any error.
func MustAnyRaw(x *types.Any) any {
	res, err := AnyRaw(x)
	if err != nil {
		panic(err.Error())
	}

	return res
}

// AnyRaw get underlying wrapped value within types.Any.
func AnyRaw(x *types.Any) (any, error) {
	if x == nil {
		return nil, fmt.Errorf("nil value")
	}

	switch x.TypeUrl {
	case ArrayFieldType:
		var arr Array
		if err := proto.Unmarshal(x.Value, &arr); err != nil {
			return nil, err
		}

		if err := checkMixTypedArray(&arr); err != nil {
			return nil, err
		}

		if !EnableMixedArrayField {
			if len(arr.Arr) == 0 {
				return nil, nil
			}

			// NOTE: detect element's type and return the same []<type>{} array.
			// i.e., element in array are int64, then we get []int64{} array.
			switch arr.Arr[0].GetX().(type) {
			case *BasicTypes_I:
				res := make([]int64, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetI())
				}
				return res, nil
			case *BasicTypes_U:
				res := make([]uint64, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetU())
				}
				return res, nil
			case *BasicTypes_F:
				res := make([]float64, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetF())
				}
				return res, nil
			case *BasicTypes_B:
				res := make([]bool, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetB())
				}
				return res, nil
			case *BasicTypes_D:
				res := make([][]byte, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetD())
				}
				return res, nil
			case *BasicTypes_S:
				res := make([]string, 0, len(arr.Arr))
				for _, v := range arr.Arr {
					res = append(res, v.GetS())
				}
				return res, nil
			default:
				return nil, fmt.Errorf("unknown type %q within array", reflect.TypeOf(arr.Arr[0].GetX()).String())
			}
		} else {
			// for mix-typed element in array, we just return the []any{}.
			var res []any
			for _, v := range arr.Arr {
				switch v.GetX().(type) {
				case *BasicTypes_I:
					res = append(res, v.GetI())
				case *BasicTypes_U:
					res = append(res, v.GetU())
				case *BasicTypes_F:
					res = append(res, v.GetF())
				case *BasicTypes_B:
					res = append(res, v.GetB())
				case *BasicTypes_D:
					res = append(res, v.GetD())
				case *BasicTypes_S:
					res = append(res, v.GetS())
				default:
					return nil, fmt.Errorf("unknown type %q within array", reflect.TypeOf(v.GetX()).String())
				}
			}
			return res, nil
		}

	case DictFieldType:
		if !EnableDictField {
			return nil, fmt.Errorf("dict/map field value not allowed")
		}

		var m Map
		if err := proto.Unmarshal(x.Value, &m); err != nil {
			return nil, err
		}

		res := map[string]any{}
		for k, v := range m.Map {
			switch v.GetX().(type) {
			case *BasicTypes_I:
				res[k] = v.GetI()
			case *BasicTypes_U:
				res[k] = v.GetU()
			case *BasicTypes_F:
				res[k] = v.GetF()
			case *BasicTypes_B:
				res[k] = v.GetB()
			case *BasicTypes_D:
				res[k] = v.GetD()
			case *BasicTypes_S:
				res[k] = v.GetS()
			default:
				return nil, fmt.Errorf("unknown type %q within map", reflect.TypeOf(v.GetX()).String())
			}
		}

		return res, nil

	default:
		return nil, fmt.Errorf("unknown type %q", x.TypeUrl)
	}
}

// MustNewAnyArray wrapped mix-basic-typed list into types.Any, and panic if any error.
func MustNewAnyArray(a ...any) *types.Any {
	if x, err := NewAnyArray(a...); err != nil {
		panic(err.Error())
	} else {
		return x
	}
}

// NewAnyArray wrapped mix-basic-typed list into types.Any.
func NewAnyArray(a ...any) (*types.Any, error) {
	x, err := NewArray(a...)
	if err != nil {
		return nil, err
	}

	return types.MarshalAny(x)
}

// MustNewIntArray wrapped signed int list into types.Any, and panic if any error.
func MustNewIntArray[T int8 | int16 | int | int32 | int64](i ...T) *types.Any {
	if x, err := NewIntArray(i...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// NewIntArray wrapped signed int list into types.Any.
func NewIntArray[T int8 | int16 | int | int32 | int64](i ...T) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(i)),
	}

	for _, v := range i {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{int64(v)}})
	}

	return types.MarshalAny(arr)
}

// MustNewUintArray wrapped unsigned int list into types.Any, and panic if any error.
func MustNewUintArray[T uint16 | uint | uint32 | uint64](i ...T) *types.Any {
	if x, err := NewUintArray(i...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// NewUintArray wrapped unsigned int list into types.Any.
func NewUintArray[T uint16 | uint | uint32 | uint64](i ...T) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(i)),
	}

	for _, v := range i {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{uint64(v)}})
	}

	return types.MarshalAny(arr)
}

// MustNewFloatArray wrapped float list into types.Any, and panic if any error.
func MustNewFloatArray[T float32 | float64](f ...T) *types.Any {
	if x, err := NewFloatArray(f...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// NewFloatArray wrapped float list into types.Any.
func NewFloatArray[T float32 | float64](f ...T) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(f)),
	}

	for _, v := range f {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_F{float64(v)}})
	}

	return types.MarshalAny(arr)
}

// MustNewBoolArray wrapped boolean list into types.Any, and panic if any error.
func MustNewBoolArray(b ...bool) *types.Any {
	if x, err := NewBoolArray(b...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// NewBoolArray wrapped boolean list into types.Any.
func NewBoolArray(b ...bool) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(b)),
	}

	for _, v := range b {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_B{v}})
	}

	return types.MarshalAny(arr)
}

// NewBytesArray wrapped string list into types.Any.
func NewBytesArray(bytes ...[]byte) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(bytes)),
	}

	for _, b := range bytes {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_D{b}})
	}

	return types.MarshalAny(arr)
}

// MustNewBytesArray wrapped string list into types.Any, and panic if any error.
func MustNewBytesArray(bytes ...[]byte) *types.Any {
	if x, err := NewBytesArray(bytes...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// MustNewStringArray wrapped string list into types.Any, and panic if any error.
func MustNewStringArray(s ...string) *types.Any {
	if x, err := NewStringArray(s...); err != nil {
		panic(err)
	} else {
		return x
	}
}

// NewStringArray wrapped string list into types.Any.
func NewStringArray(s ...string) (*types.Any, error) {
	arr := &Array{
		Arr: make([]*BasicTypes, 0, len(s)),
	}

	for _, v := range s {
		arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_S{v}})
	}

	return types.MarshalAny(arr)
}

// NewArray create array value that can be used in point field.
// The types within ents can be mixed basic types.
func NewArray(ents ...any) (arr *Array, err error) {
	arr = &Array{
		Arr: make([]*BasicTypes, 0, len(ents)),
	}

	for idx, v := range ents {
		switch x := v.(type) {
		case int8:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{int64(x)}})
		case uint8:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{uint64(x)}})
		case int16:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{int64(x)}})
		case uint16:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{uint64(x)}})
		case int32:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{int64(x)}})
		case uint32:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{uint64(x)}})
		case int:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{int64(x)}})
		case uint:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{uint64(x)}})
		case int64:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_I{x}})
		case uint64:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_U{x}})
		case float64:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_F{x}})
		case float32:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_F{float64(x)}})
		case string:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_S{x}})
		case []byte:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_D{x}})
		case bool:
			arr.Arr = append(arr.Arr, &BasicTypes{X: &BasicTypes_B{x}})
		default:
			if x == nil {
				// gogoproto any do not support nil value.
				return nil, fmt.Errorf("got nil(index: %d) within array", idx)
			} else {
				return nil, fmt.Errorf("unknown element type %q(index: %d) within array",
					reflect.TypeOf(v).String(), idx)
			}
		}
	}

	if err := checkMixTypedArray(arr); err != nil {
		return nil, err
	}

	return arr, nil
}

func checkMixTypedArray(arr *Array) error {
	if !EnableMixedArrayField && len(arr.Arr) > 1 { // check if array elements are same type
		firstElemType := reflect.TypeOf(arr.Arr[0].X).String()
		for idx, elem := range arr.Arr[1:] {
			if et := reflect.TypeOf(elem.X).String(); et != firstElemType {
				return fmt.Errorf("mixed type elements found in array([0]: %s <> [%d]: %s)",
					firstElemType, idx+1, et)
			}
		}
	}

	return nil
}

// MustNewMap create map value that can be used in point field, and panic if any error.
func MustNewMap(ents map[string]any) *Map {
	x, err := NewMap(ents)
	if err != nil {
		panic(err.Error())
	}

	return x
}

// NewMap create map value that can be used in point field.
func NewMap(ents map[string]any) (dict *Map, err error) {
	if !EnableDictField {
		return nil, fmt.Errorf("dict/map field value not allowed")
	}

	dict = &Map{
		Map: map[string]*BasicTypes{},
	}

	for k, v := range ents {
		switch x := v.(type) {
		case int8:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_I{int64(x)}}
		case uint8:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_U{uint64(x)}}
		case int16:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_I{int64(x)}}
		case uint16:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_U{uint64(x)}}
		case int32:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_I{int64(x)}}
		case uint32:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_U{uint64(x)}}
		case int:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_I{int64(x)}}
		case uint:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_U{uint64(x)}}
		case int64:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_I{x}}
		case uint64:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_U{x}}
		case float64:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_F{x}}
		case float32:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_F{float64(x)}}
		case string:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_S{x}}
		case []byte:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_D{x}}
		case bool:
			dict.Map[k] = &BasicTypes{X: &BasicTypes_B{x}}
		case nil:
			dict.Map[k] = nil
		default: // value ignored
			return nil, fmt.Errorf("unknown type %q within map", reflect.TypeOf(v).String())
		}
	}

	return dict, nil
}

// NewAny create types.Any based on exist proto message.
func NewAny(x proto.Message) (*types.Any, error) {
	return types.MarshalAny(x)
}

// MustNewAny create anypb based on exist proto message, and panic if any error.
func MustNewAny(x proto.Message) *types.Any {
	if a, err := types.MarshalAny(x); err != nil {
		panic(err.Error())
	} else {
		return a
	}
}
