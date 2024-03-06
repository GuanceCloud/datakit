// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"strings"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

// Attributes binding to resource.
const (
	otelResourceServiceKey = "service.name"
)

//nolint:deadcode,unused,varcheck
const (
	// HTTP.
	otelHTTPSchemeKey = "http.scheme"
	otelHTTPMethodKey = "http.method"
	// database.
	otelDBSystemKey = "db.system"
	// message queue.
	otelMessagingSystemKey = "messaging.system"
	// rpc system.
	otelRPCSystemKey = "rpc.system"
)

const (
	ExceptionEventName     = "exception"
	ExceptionTypeKey       = "exception.type"
	ExceptionMessageKey    = "exception.message"
	ExceptionStacktraceKey = "exception.stacktrace"
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
}

func AddCustomTags(customTags []string) {
	for _, tag := range customTags {
		OTELAttributes[tag] = strings.ReplaceAll(tag, ".", "_")
	}
}
