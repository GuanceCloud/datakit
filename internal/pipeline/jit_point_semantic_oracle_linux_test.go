// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gogo/protobuf/types"
)

// A protobuf map has no serialization order. Decode only this known Any type;
// retain scalar/list order, field types and metadata, unknown fields and every
// other Point comparison. Never canonicalize arbitrary JSON strings.
func equalJSONOraclePoints(actual, expected *point.Point) (bool, string) {
	if len(actual.Fields()) != len(expected.Fields()) || len(actual.Tags()) != len(expected.Tags()) {
		return false, "field/tag count differs"
	}
	if actual.Time().UnixNano() != expected.Time().UnixNano() {
		return false, "timestamp differs"
	}
	var decodedKeys []string
	for _, field := range actual.Fields() {
		payload := field.GetA()
		if payload == nil || payload.TypeUrl != "type.googleapis.com/point.Map" {
			continue
		}
		other := expected.Fields().Get(field.Key)
		if other == nil || other.GetA() == nil {
			return false, fmt.Sprintf("map field %q missing or wrong type", field.Key)
		}
		leftField, rightField := *field, *other
		leftField.Val, rightField.Val = nil, nil
		leftAny, rightAny := *payload, *other.GetA()
		leftAny.Value, rightAny.Value = nil, nil
		if !reflect.DeepEqual(leftField, rightField) || !reflect.DeepEqual(leftAny, rightAny) {
			return false, fmt.Sprintf("map field %q metadata differs", field.Key)
		}
		var leftMap, rightMap point.Map
		leftErr := leftMap.Unmarshal(payload.Value)
		rightErr := rightMap.Unmarshal(other.GetA().Value)
		if leftErr != nil || rightErr != nil {
			if !reflect.DeepEqual(field, other) {
				return false, fmt.Sprintf("malformed map field %q differs", field.Key)
			}
			decodedKeys = append(decodedKeys, field.Key)
			continue
		}
		if !semanticProtoEqual(reflect.ValueOf(leftMap), reflect.ValueOf(rightMap)) {
			return false, fmt.Sprintf("map field %q typed content differs", field.Key)
		}
		decodedKeys = append(decodedKeys, field.Key)
	}
	return actual.EqualWithReason(expected, point.EqualWithoutKeys(decodedKeys...))
}

// semanticProtoEqual is deliberately limited to decoded protobuf values used
// by this oracle. It preserves exact type/shape/value comparison, except that
// corresponding NaNs compare equal so IEEE's NaN != NaN rule cannot make two
// semantically identical Point payloads fail a differential test.
func semanticProtoEqual(left, right reflect.Value) bool {
	if !left.IsValid() || !right.IsValid() {
		return left.IsValid() == right.IsValid()
	}
	if left.Type() != right.Type() {
		return false
	}
	switch left.Kind() {
	case reflect.Interface, reflect.Pointer:
		if left.IsNil() || right.IsNil() {
			return left.IsNil() == right.IsNil()
		}
		return semanticProtoEqual(left.Elem(), right.Elem())
	case reflect.Struct:
		for i := 0; i < left.NumField(); i++ {
			if !semanticProtoEqual(left.Field(i), right.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if left.Kind() == reflect.Slice && (left.IsNil() != right.IsNil()) {
			return false
		}
		if left.Len() != right.Len() {
			return false
		}
		for i := 0; i < left.Len(); i++ {
			if !semanticProtoEqual(left.Index(i), right.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if left.IsNil() != right.IsNil() || left.Len() != right.Len() {
			return false
		}
		for _, key := range left.MapKeys() {
			rightValue := right.MapIndex(key)
			if !rightValue.IsValid() || !semanticProtoEqual(left.MapIndex(key), rightValue) {
				return false
			}
		}
		return true
	case reflect.Float32, reflect.Float64:
		return left.Float() == right.Float() || math.IsNaN(left.Float()) && math.IsNaN(right.Float())
	default:
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}
}

func TestJSONOracleMapComparisonIsTypedAndOrderIndependent(t *testing.T) {
	left := newRealScriptPoint("map", map[string]any{"value": int64(0)})
	right := newRealScriptPoint("map", map[string]any{"value": int64(0)})
	left.Fields().Get("value").Val = &point.Field_A{A: &types.Any{TypeUrl: "type.googleapis.com/point.Map"}}
	right.Fields().Get("value").Val = &point.Field_A{A: &types.Any{TypeUrl: "type.googleapis.com/point.Map"}}
	// Marshal singleton maps separately to force opposite wire entry order.
	a, err := (&point.Map{Map: map[string]*point.BasicTypes{"a": {X: &point.BasicTypes_I{I: 1}}}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	b, err := (&point.Map{Map: map[string]*point.BasicTypes{"b": {X: &point.BasicTypes_I{I: 2}}}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	left.Fields().Get("value").GetA().Value = append(append([]byte{}, a...), b...)
	right.Fields().Get("value").GetA().Value = append(append([]byte{}, b...), a...)
	if equal, reason := equalJSONOraclePoints(left, right); !equal {
		t.Fatal(reason)
	}
	right.Fields().Get("value").Unit = "different"
	if equal, _ := equalJSONOraclePoints(left, right); equal {
		t.Fatal("ignored metadata")
	}
	right.Fields().Get("value").Unit = ""
	right.Fields().Get("value").GetA().Value = a
	if equal, _ := equalJSONOraclePoints(left, right); equal {
		t.Fatal("ignored missing map member")
	}
}
