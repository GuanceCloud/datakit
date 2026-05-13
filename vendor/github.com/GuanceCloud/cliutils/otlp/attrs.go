// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type StringMapOptions struct {
	KeyNormalizer func(string) string
	MaxValueLen   int
	DropKeys      map[string]struct{}
}

// FindAttribute returns the first attribute with key and its position.
func FindAttribute(attrs []*common.KeyValue, key string) (*common.KeyValue, int, bool) {
	for idx, attr := range attrs {
		if isNil(attr) {
			continue
		}
		if attr.GetKey() == key {
			return attr, idx, true
		}
	}

	return nil, -1, false
}

// AnyValueToInterface converts an OTLP AnyValue into a Go value.
func AnyValueToInterface(value *common.AnyValue) (any, bool) {
	if isNil(value) {
		return nil, false
	}

	switch anyValueFieldName(value) {
	case "bytes_value":
		return value.GetBytesValue(), true
	case "string_value":
		return value.GetStringValue(), true
	case "double_value":
		return value.GetDoubleValue(), true
	case "int_value":
		return value.GetIntValue(), true
	case "bool_value":
		return value.GetBoolValue(), true
	case "array_value", "kvlist_value":
		bts, err := marshalAnyValueSubMessage(value)
		if err != nil {
			return nil, false
		}
		return string(bts), true
	default:
		return nil, false
	}
}

// AnyValueToString converts an OTLP AnyValue into a string form suitable for tags or fields.
func AnyValueToString(value *common.AnyValue) (string, bool) {
	if isNil(value) {
		return "", false
	}

	switch anyValueFieldName(value) {
	case "bytes_value":
		return string(value.GetBytesValue()), true
	case "string_value":
		return value.GetStringValue(), true
	case "double_value":
		return strconv.FormatFloat(value.GetDoubleValue(), 'f', -1, 64), true
	case "int_value":
		return strconv.FormatInt(value.GetIntValue(), 10), true
	case "bool_value":
		return strconv.FormatBool(value.GetBoolValue()), true
	case "array_value", "kvlist_value":
		bts, err := marshalAnyValueSubMessage(value)
		if err != nil {
			return "", false
		}
		return string(bts), true
	default:
		return "", false
	}
}

// AttributesToStringMap converts OTLP attributes to a string map.
func AttributesToStringMap(attrs []*common.KeyValue, opts StringMapOptions) map[string]string {
	out := make(map[string]string, len(attrs))

	for _, attr := range attrs {
		if isNil(attr) {
			continue
		}

		key := attr.GetKey()
		if _, ok := opts.DropKeys[key]; ok {
			continue
		}
		if opts.KeyNormalizer != nil {
			key = opts.KeyNormalizer(key)
		}

		value, ok := AnyValueToString(attr.GetValue())
		if !ok {
			continue
		}
		if opts.MaxValueLen > 0 && len(value) > opts.MaxValueLen {
			value = value[:opts.MaxValueLen]
		}

		out[key] = value
	}

	return out
}

// MergeStringMapsAsTags merges maps into point tags, replacing dots in keys.
func MergeStringMapsAsTags(maps ...map[string]string) point.KVs {
	var kvs point.KVs

	for _, m := range maps {
		for key, value := range m {
			kvs = kvs.AddTag(normalizePointKey(key), value)
		}
	}

	return kvs
}

// MergeStringMapsAsFields merges maps into point fields, replacing dots in keys.
func MergeStringMapsAsFields(maps ...map[string]string) point.KVs {
	var kvs point.KVs

	for _, m := range maps {
		for key, value := range m {
			kvs = kvs.Add(normalizePointKey(key), value)
		}
	}

	return kvs
}

func normalizePointKey(key string) string {
	return strings.ReplaceAll(key, ".", "_")
}

func anyValueFieldName(value *common.AnyValue) string {
	msg := value.ProtoReflect()
	oneof := msg.Descriptor().Oneofs().ByName("value")
	if oneof == nil {
		return ""
	}

	field := msg.WhichOneof(oneof)
	if field == nil {
		return ""
	}

	return string(field.Name())
}

func marshalAnyValueSubMessage(value *common.AnyValue) ([]byte, error) {
	msg := value.ProtoReflect()
	oneof := msg.Descriptor().Oneofs().ByName("value")
	if oneof == nil {
		return nil, nil
	}

	field := msg.WhichOneof(oneof)
	if field == nil {
		return nil, nil
	}

	sub := msg.Get(field).Message()
	if !sub.IsValid() {
		return nil, nil
	}

	bts, err := protojson.Marshal(sub.Interface())
	if err != nil {
		bts, err = json.Marshal(sub.Interface())
		if err != nil {
			return nil, err
		}
	}

	return bts, nil
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() { // nolint
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
