package opentelemetry

import (
	"strconv"

	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

/*
从 otel.span 对象解析到 datakit.span 中的时候，有些字段无法没有对应，不应当主动丢弃，暂时放进tags中

see : vendor/go.opentelemetry.io/proto/otlp/trace/v1/trace.pb.go:383
*/

var (
	DroppedAttributesCount = "dropped_attributes_count"
	DroppedEventsCount     = "dropped_events_count"
	DroppedLinksCount      = "dropped_links_count"
	Events                 = "events_count"
	Links                  = "links_count"
)

func getOtherTags(span *v1.Span) map[string]string {
	otherTags := make(map[string]string)
	if span.DroppedAttributesCount != 0 {
		count := strconv.Itoa(int(span.DroppedAttributesCount))
		otherTags[DroppedAttributesCount] = count
	}
	if span.DroppedEventsCount != 0 {
		count := strconv.Itoa(int(span.DroppedEventsCount))
		otherTags[DroppedEventsCount] = count
	}
	if span.DroppedLinksCount != 0 {
		count := strconv.Itoa(int(span.DroppedLinksCount))
		otherTags[DroppedLinksCount] = count
	}
	if len(span.Events) != 0 {
		count := strconv.Itoa(len(span.Events))
		otherTags[Events] = count
	}
	if len(span.Links) != 0 {
		count := strconv.Itoa(len(span.Links))
		otherTags[Links] = count
	}
	return otherTags
}
