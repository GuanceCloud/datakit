// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"strings"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

// Attributes binding to resource.
const (
	otelResourceServiceKey = "service.name"
	otelHTTPSchemeKey      = "http_scheme"
	otelHTTPMethodKey      = "http_method"
	otelDBSystemKey        = "db_system"
	otelMessagingSystemKey = "messaging_system"
	otelRPCSystemKey       = "rpc_system"
	defaultTraceAPI        = "/otel/v1/traces"
	defaultMetricAPI       = "/otel/v1/metrics"
	defaultLogAPI          = "/otel/v1/logs"
)

const (
	ExceptionEventName     = "exception"
	ExceptionTypeKey       = "exception.type"
	ExceptionMessageKey    = "exception.message"
	ExceptionStacktraceKey = "exception.stacktrace"
)

// Histogram 和 Summary 有一些固定的后缀和标签。
const (
	metricName   = "otel_service"
	unitTag      = "unit"
	bucketSuffix = "_bucket"
	sumSuffix    = "_sum"
	countSuffix  = "_count"
	avgSuffix    = "_avg"
	minSuffix    = "_min"
	maxSuffix    = "_max"
	leTag        = "le"
	infSuffix    = "+Inf" // 固定且大小写敏感
)

var otelErrKeyToDkErrKey = map[string]string{
	ExceptionTypeKey:       itrace.FieldErrType,
	ExceptionMessageKey:    itrace.FieldErrMessage,
	ExceptionStacktraceKey: itrace.FieldErrStack,
}

var (
	convertToZhaoShang = false
	convertToDD        = false
)

var spanKinds = map[int32]string{
	0: "unspecified",
	1: "internal",
	2: "server",
	3: "client",
	4: "producer",
	5: "consumer",
}

// OTELAttributes 公共标签，其中有版本变更的以使用最新的为准。
var OTELAttributes = map[string]string{
	// DB
	"db.system":    "db_system",
	"db.operation": "db_operation",
	"db.name":      "db_name",
	"db.statement": "db_statement",

	// common
	"server.address":       "server_address",
	"net.host.name":        "net_host_name",
	"server.port":          "server_port",
	"net.host.port":        "net_host_port",
	"network.peer.address": "network_peer_address",
	"network.peer.port":    "network_peer_port",
	"network.transport":    "network_transport",

	// HTTP
	"http.request.method":       "http_request_method",
	"http.method":               "http_method",
	"error.type":                "error_type",
	"http.response.status_code": "http_response_status_code",
	"http.status_code":          "http_status_code",
	"http.route":                "http_route",
	"http.target":               "http_target",
	"http.scheme":               "http_scheme",
	"http.url":                  "http_url",
	"url.full":                  "url_full",
	"url.scheme":                "url_scheme",
	"url.path":                  "url_path",
	"url.query":                 "url_query",
	"client.address":            "client_address",
	"client.port":               "client_port",

	// MQ
	"messaging.system":           "messaging_system",
	"messaging.operation":        "messaging_operation",
	"messaging.message.id":       "messaging_message.id",
	"messaging.destination.name": "messaging_destination.name",

	// RPC
	"rpc.service": "rpc_service",
	"rpc.system":  "rpc_system",

	// error
	"exception":            "exception",
	"exception.type":       "exception_type",
	"exception.message":    "exception_message",
	"exception.stacktrace": "exception_stacktrace",

	"container.name": "container_name",
	"process.pid":    "process_pid",
	"project":        "project",
	"version":        "version",
	"env":            "env",
	"host":           "host",
	"pod_name":       "pod_name",
	"pod_namespace":  "pod_namespace",
}

func AddCustomTags(customTags []string) {
	for _, tag := range customTags {
		OTELAttributes[tag] = strings.ReplaceAll(tag, ".", "_")
	}
}

func getServiceNameBySystem(atts []*common.KeyValue, defaultName string) string {
	for _, keyValue := range atts {
		key := keyValue.GetKey()
		if key == "db.system" || key == "rpc.system" || key == "messaging.system" {
			if system := keyValue.GetValue().GetStringValue(); system != "" {
				return system
			}
		}
	}

	return defaultName
}

// delMetricKey: 删除无效的key，节省内存空间。
var delMetricKey = []string{
	"process.command_line",
	"process.executable.path",
	"process.runtime.description",
	"process.runtime.name",
	"process.runtime.version",
	"telemetry.distro.name",
	"telemetry.distro.version",
	"telemetry.sdk.language",
	"telemetry.sdk.name",
	"telemetry.sdk.version",
}
