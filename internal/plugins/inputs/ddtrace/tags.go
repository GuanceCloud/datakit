// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace tags.
package ddtrace

import (
	"strings"
	"sync"
)

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
	"error.type":        "error_type",
	"system.pid":        "pid",
	"error.msg":         "error_message",
	"project":           "project",
	"version":           "version",
	"env":               "env",
	"host":              "host",
	"pod_name":          "pod_name",
	"pod_namespace":     "pod_namespace",
	"_dd.base_service":  "base_service",
	// db 类型
	"db.type":      "db_system",
	"db.instance":  "db_name",
	"db.operation": "db_operation",
	"out.host":     "out_host",
}

var ddTagsLock sync.RWMutex

func setCustomTags(customTags []string) {
	ddTagsLock.Lock()
	for _, tag := range customTags {
		log.Debugf("set customtag key %s to ddTags", tag)
		ddTags[tag] = strings.ReplaceAll(tag, ".", "_")
	}
	ddTagsLock.Unlock()
}
