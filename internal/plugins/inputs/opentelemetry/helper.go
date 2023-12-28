// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"regexp"
	"strings"

	"github.com/GuanceCloud/cliutils/point"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
)

type getAttributeFunc func(key string, attributes []*common.KeyValue) (*common.KeyValue, bool)

func getAttr(key string, attributes []*common.KeyValue) (*common.KeyValue, bool) {
	for _, attr := range attributes {
		if attr.Key == key {
			return attr, true
		}
	}

	return nil, false
}

func getAttrWrapper(ignore []*regexp.Regexp) getAttributeFunc {
	if len(ignore) == 0 {
		return getAttr
	} else {
		return func(key string, attributes []*common.KeyValue) (*common.KeyValue, bool) {
			for _, rexp := range ignore {
				if rexp.MatchString(key) {
					return nil, false
				}
			}

			return getAttr(key, attributes)
		}
	}
}

type extractAttributesFunc func(src []*common.KeyValue) (dest []*common.KeyValue)

func extractAttrs(src []*common.KeyValue) (dest []*common.KeyValue) {
	dest = append(dest, src...)

	return
}

func extractAttrsWrapper(ignore []*regexp.Regexp) extractAttributesFunc {
	if len(ignore) == 0 {
		return extractAttrs
	} else {
		return func(src []*common.KeyValue) (dest []*common.KeyValue) {
		NEXT_ATTR:
			for _, v := range src {
				for _, rexp := range ignore {
					if rexp.MatchString(v.Key) {
						continue NEXT_ATTR
					}
				}
				dest = append(dest, v)
			}

			return
		}
	}
}

func newAttributes(attrs []*common.KeyValue) *attributes {
	a := &attributes{}
	a.attrs = append(a.attrs, attrs...)

	return a
}

type attributes struct {
	attrs []*common.KeyValue
}

// nolint: deadcode,unused
func (a *attributes) loop(proc func(i int, k string, v *common.KeyValue) bool) {
	for i, v := range a.attrs {
		if !proc(i, v.Key, v) {
			break
		}
	}
}

func (a *attributes) merge(attrs ...*common.KeyValue) *attributes {
	for _, v := range attrs {
		if _, i := a.find(v.Key); i != -1 {
			a.attrs[i] = v
		} else {
			a.attrs = append(a.attrs, v)
		}
	}

	return a
}

func (a *attributes) find(key string) (*common.KeyValue, int) {
	for i := len(a.attrs) - 1; i >= 0; i-- {
		if a.attrs[i].Key == key {
			return a.attrs[i], i
		}
	}

	return nil, -1
}

func (a *attributes) splite() (map[string]string, map[string]interface{}) {
	shadowTags := make(map[string]string)
	metrics := make(map[string]interface{})
	if len(a.attrs) > 100 {
		a.attrs = a.attrs[:100]
	}
	for _, v := range a.attrs {
		key := strings.ReplaceAll(v.Key, ".", "_")
		switch v.Value.Value.(type) {
		case *common.AnyValue_BytesValue, *common.AnyValue_StringValue:
			if s := v.Value.GetStringValue(); len(s) > 1024 {
				metrics[key] = s
			} else {
				shadowTags[key] = s
			}
		case *common.AnyValue_DoubleValue:
			metrics[key] = v.Value.GetDoubleValue()
		case *common.AnyValue_IntValue:
			metrics[key] = v.Value.GetIntValue()
		}
	}

	return shadowTags, metrics
}

func attributesToKVS(spanKV point.KVs, otherAttrs, atts []*common.KeyValue) (point.KVs, []*common.KeyValue) {
	for _, v := range atts {
		if replaceKey, ok := OTELAttributes[v.Key]; ok {
			switch v.Value.Value.(type) {
			case *common.AnyValue_BytesValue, *common.AnyValue_StringValue:
				if s := v.Value.GetStringValue(); len(s) > 1024 {
					spanKV = spanKV.Add(replaceKey, s, false, true)
				} else {
					spanKV = spanKV.MustAddTag(replaceKey, s)
				}
			case *common.AnyValue_DoubleValue:
				spanKV = spanKV.Add(replaceKey, v.Value.GetDoubleValue(), false, true)
			case *common.AnyValue_IntValue:
				spanKV = spanKV.Add(replaceKey, v.Value.GetIntValue(), false, true)
			}
		} else {
			otherAttrs = append(otherAttrs, v)
		}
	}
	return spanKV, otherAttrs
}
