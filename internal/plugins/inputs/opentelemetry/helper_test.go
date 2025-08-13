// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"testing"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
)

func TestInput_selectAttrs(t *testing.T) {
	ipt := &Input{commonAttrs: map[string]string{}, CustomerTagsAll: true}

	atts := make([]*common.KeyValue, 0)

	arr := &common.KeyValue{
		Key: "test_array",
		Value: &common.AnyValue{
			Value: &common.AnyValue_ArrayValue{
				ArrayValue: &common.ArrayValue{
					Values: []*common.AnyValue{
						{
							Value: &common.AnyValue_StringValue{StringValue: "javaagent:/ddjava-agent.jar"},
						},
						{
							Value: &common.AnyValue_StringValue{StringValue: "-Ddd.service.name=tmall"},
						},
						{
							Value: &common.AnyValue_StringValue{StringValue: "-jar tmall.jar"},
						},
					},
				},
			},
		},
	}
	atts = append(atts, arr)

	kvList := &common.KeyValue{
		Key: "test_kvlist",
		Value: &common.AnyValue{
			Value: &common.AnyValue_KvlistValue{
				KvlistValue: &common.KeyValueList{
					Values: []*common.KeyValue{
						{
							Key: "version",
							Value: &common.AnyValue{
								Value: &common.AnyValue_StringValue{StringValue: "1.0.1"},
							},
						},
						{
							Key: "env",
							Value: &common.AnyValue{
								Value: &common.AnyValue_StringValue{StringValue: "prod"},
							},
						},
					},
				},
			},
		},
	}
	atts = append(atts, kvList)

	ipt.jmarshaler = &protojsonMarshaler{}
	kvs, _ := ipt.selectAttrs(atts)
	f := kvs.Get("test_array")
	t.Log(f.GetS())

	f = kvs.Get("test_kvlist")
	t.Log(f.GetS())
	t.Logf("")

	ipt.jmarshaler = &jsoniterMarshaler{}
	kvs, _ = ipt.selectAttrs(atts)
	f = kvs.Get("test_array")

	t.Log(f.GetS())

	f = kvs.Get("test_kvlist")
	t.Log(f.GetS())
}
