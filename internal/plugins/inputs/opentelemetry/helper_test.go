// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestInput_selectAttrs(t *testing.T) {
	ipt := &Input{commonAttrs: map[string]string{}, CustomerTagsAll: true, customTagsX: itrace.NewCustomTags([]string{}, otelPubAttrs)}

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

func Test_getDBHost(t *testing.T) {
	type args struct {
		atts []*common.KeyValue
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "db",
			args: args{
				atts: []*common.KeyValue{
					{
						Key: "db.system",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "mysql"},
						},
					},
					{
						Key: "language",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "java"},
						},
					},
					{
						Key: "server.address",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "localhost"},
						},
					},
					nil,
					nil,
				},
			},
			want: "localhost",
		},
		{
			name: "empty",
			args: args{
				atts: []*common.KeyValue{
					{
						Key: "db.name",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "mysql"},
						},
					},
					{
						Key: "language",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "java"},
						},
					},
					{
						Key: "server.address",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{StringValue: "localhost"},
						},
					},
					nil,
					nil,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getDBHost(tt.args.atts), "getDBHost(%v)", tt.args.atts)
		})
	}
}
