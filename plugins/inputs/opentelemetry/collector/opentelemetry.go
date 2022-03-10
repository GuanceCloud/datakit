package collector

// see:vendor/go.opentelemetry.io/otel/semconv/v1.4.0/trace.go.
const (
	otelResourceServiceKey = "service.name"
	defaultServiceVal      = "unknown.service"
	// HTTP.
	otelResourceHTTPMethodKey     = "http.method"
	otelResourceHTTPStatusCodeKey = "http.status_code"
	otelResourceContainerNameKey  = "container.name"
	otelResourceProcessPidKey     = "process.pid"

	// 从 otel.span 对象解析到 datakit.span 中的时候，有些字段无法没有对应，不应当主动丢弃，暂时放进tags中
	// see : vendor/go.opentelemetry.io/proto/otlp/trace/v1/trace.pb.go:383.
	DroppedAttributesCount = "dropped_attributes_count"
	DroppedEventsCount     = "dropped_events_count"
	DroppedLinksCount      = "dropped_links_count"
	Events                 = "events_count"
	Links                  = "links_count"
)
