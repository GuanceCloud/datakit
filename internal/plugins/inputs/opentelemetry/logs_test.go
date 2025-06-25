// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry logs testing.
package opentelemetry

import (
	"encoding/hex"
	"testing"

	v11 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	logs "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/logs/v1"
	v1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
	"github.com/stretchr/testify/assert"
)

func TestParseLogsRequest(t *testing.T) {
	traceID, _ := hex.DecodeString("818616084f850520843d19e3936e4720")
	spanID, _ := hex.DecodeString("843d19e3936e4720")

	type args struct {
		resourceLogss []*logs.ResourceLogs
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "otel-logs",
			args: args{resourceLogss: []*logs.ResourceLogs{
				{
					Resource: &v1.Resource{
						Attributes: []*v11.KeyValue{
							{
								Key: "service.name",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{StringValue: "tmall"},
								},
							},
							{
								Key: "version",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{StringValue: "v1.0.1"},
								},
							},
						},
						DroppedAttributesCount: 0,
					},
					ScopeLogs: []*logs.ScopeLogs{
						{
							Scope: &v11.InstrumentationScope{
								Attributes: []*v11.KeyValue{
									{
										Key: "jvm",
										Value: &v11.AnyValue{
											Value: &v11.AnyValue_StringValue{StringValue: "jdk-11"},
										},
									},
								},
							},
							LogRecords: []*logs.LogRecord{
								{
									TimeUnixNano:         0,
									ObservedTimeUnixNano: 0,
									SeverityNumber:       logs.SeverityNumber_SEVERITY_NUMBER_INFO,
									SeverityText:         "INFO",
									Body: &v11.AnyValue{
										Value: &v11.AnyValue_StringValue{StringValue: "this message"},
									},
									Attributes:             nil,
									DroppedAttributesCount: 0,
									Flags:                  0,
									TraceId:                traceID,
									SpanId:                 spanID,
								},
							},
							SchemaUrl: "",
						},
					},
					SchemaUrl: "otel schema_url",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()

			pts := ipt.parseLogRequest(tt.args.resourceLogss)
			for _, pt := range pts {
				assert.Equal(t, pt.Get("service"), "tmall")
				assert.Equal(t, pt.Get("message"), "this message")
				assert.Equal(t, pt.Get("jvm"), "jdk-11")
				assert.Equal(t, pt.Get("version"), "v1.0.1")
				assert.Equal(t, pt.Get("trace_id"), "818616084f850520843d19e3936e4720")
				assert.Equal(t, pt.Get("span_id"), "843d19e3936e4720")
				assert.Equal(t, pt.Get("status"), "info")
			}
		})
	}
}

func Test_splitByByteLength(t *testing.T) {
	type args struct {
		s      string
		length int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "case 1",
			args: args{
				s:      "1234567890",
				length: 10,
			},
			want: []string{"1234567890"},
		},
		{
			name: "case 2",
			args: args{
				s:      "测试中文和符号*&，。、%切割",
				length: 7,
			},
			want: []string{"测试中文和符号", "*&，。、%切", "割"},
		},
		{
			name: "case 3",
			args: args{
				s:      "1234567890",
				length: 8,
			},
			want: []string{"12345678", "90"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, splitByByteLength(tt.args.s, tt.args.length), "splitByByteLength(%v, %v)", tt.args.s, tt.args.length)
		})
	}
}
