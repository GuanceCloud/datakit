// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"

// Attributes binding to resource
const (
	otelResourceServiceKey        = "service.name"
	otelResourceServiceVersionKey = "service.version"
	otelResourceContainerNameKey  = "container.name"
	otelResourceProcessIDKey      = "process.pid"
)

// Attributes binding to instrument

// Attributes binding to span
const (
	// HTTP.
	otelHTTPSchemeKey     = "http.scheme"
	otelHTTPMethodKey     = "http.method"
	otelHTTPStatusCodeKey = "http.status_code"
	// 从 otel.span 对象解析到 datakit.span 中的时候，有些字段无法没有对应，不应当主动丢弃，暂时放进tags中
	// see : vendor/go.opentelemetry.io/proto/otlp/trace/v1/trace.pb.go:383.
	DroppedAttributesCount = "dropped_attributes_count"
	DroppedEventsCount     = "dropped_events_count"
	DroppedLinksCount      = "dropped_links_count"
	Events                 = "events_count"
	Links                  = "links_count"
	// db
	otelDBSystemKey = "db.system"
)

// Attributes binding to event

const (
	ExceptionEventName     = "exception"
	ExceptionTypeKey       = "exception.type"
	ExceptionMessageKey    = "exception.message"
	ExceptionStacktraceKey = "exception.stacktrace"
)

var otelErrKeyToDkErrKey = map[string]string{
	ExceptionTypeKey:       itrace.TAG_ERR_TYPE,
	ExceptionMessageKey:    itrace.TAG_ERR_MESSAGE,
	ExceptionStacktraceKey: itrace.TAG_ERR_STACK,
}
