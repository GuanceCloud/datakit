// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	commonpb "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/compiled/v1/common"
)

type getAttributeFunc func(key string, attributes []*commonpb.KeyValue) (*commonpb.KeyValue, bool)

func getAttr(key string, attributes []*commonpb.KeyValue) (*commonpb.KeyValue, bool) {
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
		return func(key string, attributes []*commonpb.KeyValue) (*commonpb.KeyValue, bool) {
			for _, rexp := range ignore {
				if rexp.MatchString(key) {
					return nil, false
				}
			}

			return getAttr(key, attributes)
		}
	}
}

type extractAttributesFunc func(attributes []*commonpb.KeyValue) (tags map[string]string, fields map[string]interface{})

func extractAttr(attributes []*commonpb.KeyValue) (tags map[string]string, fields map[string]interface{}) {
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	for _, attr := range attributes {
		switch attr.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue, *commonpb.AnyValue_BytesValue:
			if v := attr.Value.GetStringValue(); len(v) > point.MaxTagValueLen {
				fields[attr.Key] = v
			} else {
				tags[attr.Key] = v
			}
		case *commonpb.AnyValue_IntValue:
			fields[attr.Key] = attr.Value.GetIntValue()
		case *commonpb.AnyValue_DoubleValue:
			fields[attr.Key] = attr.Value.GetDoubleValue()
		default:
			continue
		}
	}

	return
}

func extractAttrWrapper(ignore []*regexp.Regexp) extractAttributesFunc {
	if len(ignore) == 0 {
		return extractAttr
	} else {
		return func(attributes []*commonpb.KeyValue) (tags map[string]string, fields map[string]interface{}) {
			tags = make(map[string]string)
			fields = make(map[string]interface{})
		NEXT_ATTR:
			for _, attr := range attributes {
				for _, rexp := range ignore {
					if rexp.MatchString(attr.Key) {
						continue NEXT_ATTR
					}
				}

				switch attr.Value.Value.(type) {
				case *commonpb.AnyValue_StringValue, *commonpb.AnyValue_BytesValue:
					if v := attr.Value.GetStringValue(); len(v) > point.MaxTagValueLen {
						fields[attr.Key] = v
					} else {
						tags[attr.Key] = v
					}
				case *commonpb.AnyValue_IntValue:
					fields[attr.Key] = attr.Value.GetIntValue()
				case *commonpb.AnyValue_DoubleValue:
					fields[attr.Key] = attr.Value.GetDoubleValue()
				default:
					continue
				}
			}

			return
		}
	}
}
