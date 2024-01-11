// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace for daoke
package trace

import (
	"testing"
)

var testSpanStr = `
{
	"traceId": "q9f9RaeqewMpQKWjWzj1Eg==",
	"spanId": "1f2n/O2U20w=",
	"operationName": "BaseController.readiness",
	"references": [{
		"traceId": "q9f9RaeqewMpQKWjWzj1Eg==",
		"spanId": "oD09Zb3lN1o="
	}],
	"startTime": "2023-12-01T10:47:45.853375888Z",
	"duration": "0.001245672s",
	"tags": [{
		"key": "thread.id",
		"vType": "INT64",
		"vInt64": "88"
	}, {
		"key": "thread.name",
		"vStr": "http-nio-9001-exec-8"
	}, {
		"key": "otel.library.name",
		"vStr": "io.opentelemetry.spring-webmvc-3.1"
	}, {
		"key": "otel.library.version",
		"vStr": "1.10.0"
	}, {
		"key": "internal.span.format",
		"vStr": "proto"
	}],
	"process": {
		"serviceName": "SUBSYS_ADGW_COREGW",
		"tags": [{
			"key": "clustername",
			"vStr": "zsc-k8sca"
		}, {
			"key": "hostname",
			"vStr": "adgw-coregw-py-6cbfcc9678-g2pmb"
		}, {
			"key": "ip",
			"vStr": "149.162.10.0"
		}, {
			"key": "jaeger.version",
			"vStr": "opentelemetry-java"
		}, {
			"key": "profile.env",
			"vStr": "zsc"
		}, {
			"key": "service.name",
			"vStr": "SUBSYS_ADGW_COREGW"
		}, {
			"key": "servicegroup",
			"vStr": "py"
		}, {
			"key": "telemetry.sdk.language",
			"vStr": "java"
		}, {
			"key": "telemetry.sdk.name",
			"vStr": "opentelemetry"
		}, {
			"key": "telemetry.sdk.version",
			"vStr": "1.10.0"
		}]
	}
}`

func TestParseToJaeger(t *testing.T) {
	//nolint

	pts, err := ParseToJaeger([]byte(testSpanStr))
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	t.Logf("point=%s", pts.LineProto())
}

func BenchmarkParseToJaeger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseToJaeger([]byte(testSpanStr))
		if err != nil {
			b.Errorf("err=%v", err)
		}
	}
}

func Test_decoder(t *testing.T) {
	type args struct {
		bs64 string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "case_1", args: args{bs64: "q9f9RaeqewMpQKWjWzj1Eg=="}, want: ""},
		{name: "case_1", args: args{bs64: "1f2n/O2U20w="}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decoder(tt.args.bs64); got != tt.want {
				t.Logf("decoder() = %v", got)
			}
		})
	}
}
