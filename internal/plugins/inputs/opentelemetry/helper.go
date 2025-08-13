// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"strings"

	"github.com/GuanceCloud/cliutils/point"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
)

func getAttr(key string, attributes []*common.KeyValue) (*common.KeyValue, int) {
	for idx, attr := range attributes {
		if attr == nil {
			continue
		}

		if attr.Key == key {
			return attr, idx
		}
	}

	return nil, -1
}

// selectAttrs extract common attrs as kvs, non-common attrs are merged.
func (ipt *Input) selectAttrs(atts []*common.KeyValue) (kvs point.KVs, merged []*common.KeyValue) {
	for _, v := range atts {
		if v == nil { // the attribute may have been dropped by ipt.CleanMessage
			continue
		}

		replaceKey, ok := ipt.commonAttrs[v.Key]

		if ipt.CustomerTagsAll {
			if replaceKey == "" {
				replaceKey = strings.ReplaceAll(v.Key, ".", "_")
			}
		} else {
			if !ok {
				merged = append(merged, v)
				continue
			}
		}

		// else
		switch v.Value.Value.(type) {
		case *common.AnyValue_BytesValue,
			*common.AnyValue_StringValue:
			if s := v.Value.GetStringValue(); len(s) > 1024 { // len(tag-value) should <= 1024
				kvs = kvs.Set(replaceKey, s) // and add it in field
			} else {
				kvs = kvs.SetTag(replaceKey, s)
			}
		case *common.AnyValue_DoubleValue:
			kvs = kvs.Set(replaceKey, v.Value.GetDoubleValue())
		case *common.AnyValue_IntValue:
			kvs = kvs.Set(replaceKey, v.Value.GetIntValue())
		case *common.AnyValue_BoolValue:
			kvs = kvs.Set(replaceKey, v.Value.GetBoolValue())
		case *common.AnyValue_KvlistValue:
			kvs = kvs.Set(replaceKey, v.Value.GetKvlistValue().String())
		case *common.AnyValue_ArrayValue:
			kvs = kvs.Set(replaceKey, v.Value.GetArrayValue().String())
		default: // passed
		}
	}

	return kvs, merged
}

func getDBHost(atts []*common.KeyValue) string {
	var isDB bool
	for _, v := range atts {
		if v == nil {
			continue
		}

		if v.Key == "db.system" {
			isDB = true
			break
		}
	}

	if !isDB {
		return ""
	}

	for _, attr := range atts {
		if attr.Key == "net.peer.name" || attr.Key == "server.address" {
			return attr.Value.GetStringValue()
		}
	}
	return ""
}
