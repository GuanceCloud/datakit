package opentelemetry

import itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"

// see:vendor/go.opentelemetry.io/otel/semconv/v1.4.0/trace.go.
const (
	otelResourceServiceKey        = "service.name"
	otelResourceServiceVersionKey = "service.version"
	otelServiceName               = "otel-service"
	otelUnknownServiceName        = "unknown_service"
	// HTTP.
	otelHTTPSchemeKey            = "http.scheme"
	otelHTTPMethodKey            = "http.method"
	otelHTTPStatusCodeKey        = "http.status_code"
	otelResourceContainerNameKey = "container.name"
	otelResourceProcessIDKey     = "process.pid"
	// 从 otel.span 对象解析到 datakit.span 中的时候，有些字段无法没有对应，不应当主动丢弃，暂时放进tags中
	// see : vendor/go.opentelemetry.io/proto/otlp/trace/v1/trace.pb.go:383.
	DroppedAttributesCount = "dropped_attributes_count"
	DroppedEventsCount     = "dropped_events_count"
	DroppedLinksCount      = "dropped_links_count"
	Events                 = "events_count"
	Links                  = "links_count"
	// for exception event.
	ExceptionEventName     = "exception"
	ExceptionTypeKey       = "exception.type"
	ExceptionMessageKey    = "exception.message"
	ExceptionStacktraceKey = "exception.stacktrace"
	// db
	otelDBSystemKey = "db.system"
)

var otelErrKeyToDkErrKey = map[string]string{
	ExceptionTypeKey:       itrace.TAG_ERR_TYPE,
	ExceptionMessageKey:    itrace.TAG_ERR_MESSAGE,
	ExceptionStacktraceKey: itrace.TAG_ERR_STACK,
}
