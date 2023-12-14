// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace tags.
package ddtrace

// ddTags DDtrace.
var ddTags = map[string]string{
	"http.url":          "http_url",
	"http.hostname":     "http_hostname",
	"http.route":        "http_route",
	"http.status_code":  "http_status_code",
	"http.method":       "http_method",
	"http.client_ip":    "http_client_ip",
	"sampling.priority": "sampling_priority",
	"span.kind":         "span_kind",
	"error":             "error",
	"runtime.name":      "runtime_name",
	"dd.version":        "dd_version",
	"dd.env":            "dd_env",
	"error.message":     "error_message",
	"error.stack":       "error_stack",
	"error_type":        "error_type",
	"system.pid":        "pid",
	"error.msg":         "error_message",
	"project":           "project",
	"version":           "version",
	"env":               "env",
	"_dd.base_service":  "base_service",
}
