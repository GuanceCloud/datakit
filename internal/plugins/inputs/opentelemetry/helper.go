// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"github.com/GuanceCloud/cliutils/point"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
)

func getAttr(key string, attributes []*common.KeyValue) (*common.KeyValue, bool) {
	for _, attr := range attributes {
		if attr.Key == key {
			return attr, true
		}
	}

	return nil, false
}

func (ipt *Input) attributesToKVS(atts []*common.KeyValue) (spanKV point.KVs, otherAttrs []*common.KeyValue) {
	for _, v := range atts {
		if replaceKey, ok := ipt.commonAttrs[v.Key]; ok {
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

func getDBHost(atts []*common.KeyValue) string {
	var isDB bool
	for _, v := range atts {
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
