package lightpanda

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/network"
)

var traceIDKeys = map[string]struct{}{
	"traceid":            {},
	"trace-id":           {},
	"traceparent":        {},
	"x-trace-id":         {},
	"x-traceid":          {},
	"x-b3-traceid":       {},
	"x-datadog-trace-id": {},
	"dd-trace-id":        {},
	"uber-trace-id":      {},
}

func extractTraceID(rawURL string, headerSets ...network.Headers) string {
	for _, headers := range headerSets {
		if traceID := extractTraceIDFromHeaders(headers); traceID != "" {
			return traceID
		}
	}
	return extractTraceIDFromURL(rawURL)
}

func extractTraceIDFromHeaders(headers network.Headers) string {
	for key, value := range headers {
		if _, ok := traceIDKeys[normalizeTraceKey(key)]; !ok {
			continue
		}
		if traceID := normalizeTraceValue(key, headerString(value)); traceID != "" {
			return traceID
		}
	}
	return ""
}

func extractTraceIDFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for key, values := range parsed.Query() {
		if _, ok := traceIDKeys[normalizeTraceKey(key)]; !ok {
			continue
		}
		for _, value := range values {
			if traceID := normalizeTraceValue(key, value); traceID != "" {
				return traceID
			}
		}
	}
	return ""
}

func normalizeTraceKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "_", "-")
	return key
}

func normalizeTraceValue(key string, value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	if value == "" {
		return ""
	}
	if index := strings.Index(value, ","); index > 0 {
		value = strings.TrimSpace(value[:index])
	}
	switch normalizeTraceKey(key) {
	case "traceparent":
		return traceIDFromTraceparent(value)
	case "uber-trace-id":
		parts := strings.Split(value, ":")
		if len(parts) > 0 {
			return cleanTraceID(parts[0])
		}
		return ""
	default:
		return cleanTraceID(value)
	}
}

func traceIDFromTraceparent(value string) string {
	parts := strings.Split(value, "-")
	if len(parts) >= 2 {
		return cleanTraceID(parts[1])
	}
	return cleanTraceID(value)
}

func cleanTraceID(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, "'")
	return value
}

func headerString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []string:
		return strings.Join(typed, ",")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprint(value)
	}
}
