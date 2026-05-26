// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

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
	metricName = "otel_service"
)

var (
	maxLogMetricFiledLen = 1024 * 32

	spanKinds = map[int32]string{
		0: "unspecified",
		1: "internal",
		2: "server",
		3: "client",
		4: "producer",
		5: "consumer",
	}

	// otelPubAttrs 公共标签，其中有版本变更的以使用最新的为准。
	otelPubAttrs = map[string]string{
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
		"http.request.method":       "http_method", // V2 版本重大变更。
		"http.method":               "http_method",
		"error.type":                "error_type",
		"http.response.status_code": "http_status_code", // V2 版本重大变更。
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

		"telemetry.sdk.language": "sdk_language",
		"telemetry.sdk.name":     "sdk_name",
		"telemetry.sdk.version":  "sdk_version",
	}

	// delMetricKey: 删除无效的key，节省内存空间。
	delMetricKey = []string{
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
)
