//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	headerTraceparentString       = "traceparent"
	headerDatadogTraceIDString    = "x-datadog-trace-id"
	headerDatadogParentIDString   = "x-datadog-parent-id"
	headerDatadogSpanIDString     = "x-datadog-span-id"
	headerB3String                = "b3"
	headerB3TraceIDString         = "x-b3-traceid"
	headerB3SpanIDString          = "x-b3-spanid"
	headerB3ParentSpanIDString    = "x-b3-parentspanid"
	headerSW8String               = "sw8"
	headerJaegerTraceIDString     = "uber-trace-id"
	headerOpenTracerTraceIDString = "ot-tracer-traceid"
	headerOpenTracerSpanIDString  = "ot-tracer-spanid"
	headerXRayTraceIDString       = "x-amzn-trace-id"
)

const (
	maxDatadogDecimalIDLen = 20

	maxSW8TraceIDLen   = 128
	maxSW8SegmentIDLen = 128
	maxSW8SpanIDLen    = maxDatadogDecimalIDLen
)

var (
	headerTraceparent       = []byte(headerTraceparentString)
	headerDatadogTraceID    = []byte(headerDatadogTraceIDString)
	headerDatadogParentID   = []byte(headerDatadogParentIDString)
	headerDatadogSpanID     = []byte(headerDatadogSpanIDString)
	headerB3                = []byte(headerB3String)
	headerB3TraceID         = []byte(headerB3TraceIDString)
	headerB3SpanID          = []byte(headerB3SpanIDString)
	headerB3ParentSpanID    = []byte(headerB3ParentSpanIDString)
	headerSW8               = []byte(headerSW8String)
	headerJaegerTraceID     = []byte(headerJaegerTraceIDString)
	headerOpenTracerTraceID = []byte(headerOpenTracerTraceIDString)
	headerOpenTracerSpanID  = []byte(headerOpenTracerSpanIDString)
	headerXRayTraceID       = []byte(headerXRayTraceIDString)
)

type traceHeaderCarrier struct {
	traceparent string

	datadogTraceID  string
	datadogParentID string
	datadogSpanID   string

	b3              string
	b3TraceID       string
	b3SpanID        string
	b3ParentSpanID  string
	sw8             string
	jaegerTraceID   string
	otTracerTraceID string
	otTracerSpanID  string
	xrayTraceID     string
}

func (c *traceHeaderCarrier) addBytes(name, value []byte) {
	value = bytes.TrimSpace(value)
	if len(value) == 0 {
		return
	}

	switch {
	case bytes.EqualFold(name, headerDatadogTraceID):
		c.datadogTraceID = string(value)
	case bytes.EqualFold(name, headerDatadogParentID):
		c.datadogParentID = string(value)
	case bytes.EqualFold(name, headerDatadogSpanID):
		c.datadogSpanID = string(value)
	case bytes.EqualFold(name, headerTraceparent):
		c.traceparent = string(value)
	case bytes.EqualFold(name, headerB3):
		c.b3 = string(value)
	case bytes.EqualFold(name, headerB3TraceID):
		c.b3TraceID = string(value)
	case bytes.EqualFold(name, headerB3SpanID):
		c.b3SpanID = string(value)
	case bytes.EqualFold(name, headerB3ParentSpanID):
		c.b3ParentSpanID = string(value)
	case bytes.EqualFold(name, headerSW8):
		c.sw8 = string(value)
	case bytes.EqualFold(name, headerJaegerTraceID):
		c.jaegerTraceID = string(value)
	case bytes.EqualFold(name, headerOpenTracerTraceID):
		c.otTracerTraceID = string(value)
	case bytes.EqualFold(name, headerOpenTracerSpanID):
		c.otTracerSpanID = string(value)
	case bytes.EqualFold(name, headerXRayTraceID):
		c.xrayTraceID = string(value)
	}
}

func (c *traceHeaderCarrier) addString(name, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	switch {
	case strings.EqualFold(name, headerDatadogTraceIDString):
		c.datadogTraceID = value
	case strings.EqualFold(name, headerDatadogParentIDString):
		c.datadogParentID = value
	case strings.EqualFold(name, headerDatadogSpanIDString):
		c.datadogSpanID = value
	case strings.EqualFold(name, headerTraceparentString):
		c.traceparent = value
	case strings.EqualFold(name, headerB3String):
		c.b3 = value
	case strings.EqualFold(name, headerB3TraceIDString):
		c.b3TraceID = value
	case strings.EqualFold(name, headerB3SpanIDString):
		c.b3SpanID = value
	case strings.EqualFold(name, headerB3ParentSpanIDString):
		c.b3ParentSpanID = value
	case strings.EqualFold(name, headerSW8String):
		c.sw8 = value
	case strings.EqualFold(name, headerJaegerTraceIDString):
		c.jaegerTraceID = value
	case strings.EqualFold(name, headerOpenTracerTraceIDString):
		c.otTracerTraceID = value
	case strings.EqualFold(name, headerOpenTracerSpanIDString):
		c.otTracerSpanID = value
	case strings.EqualFold(name, headerXRayTraceIDString):
		c.xrayTraceID = value
	}
}

func (c *traceHeaderCarrier) traceIDs() (traceID, parentID string) {
	traceID, parentID, _ = c.traceIDsWithProvider()
	return traceID, parentID
}

func (c *traceHeaderCarrier) traceIDsWithProvider() (traceID, parentID, provider string) {
	if traceID = normalizeDecimalTraceID(c.datadogTraceID); traceID != "" {
		parentID = normalizeDecimalSpanID(c.datadogParentID)
		if parentID == "" {
			parentID = normalizeDecimalSpanID(c.datadogSpanID)
		}
		return traceID, parentID, "datadog"
	}

	if traceID, parentID = parseTraceparentValue(c.traceparent); traceID != "" {
		return traceID, parentID, "w3c"
	}

	if traceID, parentID = parseB3Headers(c.b3TraceID, c.b3SpanID, c.b3ParentSpanID, c.b3); traceID != "" {
		return traceID, parentID, "b3"
	}

	if traceID, parentID = parseSW8Header(c.sw8); traceID != "" {
		return traceID, parentID, "sw8"
	}

	if traceID, parentID = parseJaegerHeader(c.jaegerTraceID); traceID != "" {
		return traceID, parentID, "jaeger"
	}

	if traceID, parentID = parseOpenTracingHeaders(c.otTracerTraceID, c.otTracerSpanID); traceID != "" {
		return traceID, parentID, "opentracing"
	}

	if traceID, parentID = parseXRayHeader(c.xrayTraceID); traceID != "" {
		return traceID, parentID, "xray"
	}

	return "", "", ""
}

func parseTraceparentHeader(value []byte) (traceID, parentID string) {
	return parseTraceparentValue(string(bytes.TrimSpace(value)))
}

func parseTraceparentValue(value string) (traceID, parentID string) {
	if strings.TrimSpace(value) == "" {
		return "", ""
	}

	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 || parts[0] == "" {
		return "", ""
	}

	traceID, ok := normalizeHexTraceID(parts[1], 32)
	if !ok {
		return "", ""
	}
	parentID, ok = normalizeHexSpanID(parts[2])
	if !ok {
		return "", ""
	}
	return traceID, parentID
}

func parseB3Headers(traceID, spanID, parentSpanID, single string) (string, string) {
	if traceID == "" || spanID == "" {
		if traceID1, spanID1 := parseB3SingleHeader(single); traceID1 != "" && spanID1 != "" {
			traceID = traceID1
			spanID = spanID1
		}
	}

	traceID, ok := normalizeHexTraceID(traceID, 16, 32)
	if !ok {
		return "", ""
	}

	parentID := normalizeOptionalHexSpanID(spanID)
	if parentID == "" {
		parentID = normalizeOptionalHexSpanID(parentSpanID)
	}
	return traceID, parentID
}

func parseB3SingleHeader(value string) (traceID, spanID string) {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" || value == "1" || strings.EqualFold(value, "d") {
		return "", ""
	}

	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func parseSW8Header(value string) (traceID, parentID string) {
	if strings.TrimSpace(value) == "" {
		return "", ""
	}

	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 4 {
		return "", ""
	}

	traceID = decodeSW8Part(parts[1], maxSW8TraceIDLen)
	if traceID == "" {
		return "", ""
	}

	parentTraceSegmentID := decodeSW8Part(parts[2], maxSW8SegmentIDLen)
	parentSpanID := normalizeSW8SpanID(parts[3])
	if parentTraceSegmentID != "" && parentSpanID != "" {
		parentID = parentTraceSegmentID + parentSpanID
	} else {
		parentID = parentSpanID
	}
	return traceID, parentID
}

func parseJaegerHeader(value string) (traceID, parentID string) {
	if strings.TrimSpace(value) == "" {
		return "", ""
	}

	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) < 2 {
		return "", ""
	}

	traceID, ok := normalizeHexTraceIDRange(parts[0], 1, 32)
	if !ok {
		return "", ""
	}
	parentID = normalizeOptionalHexSpanID(parts[1])
	return traceID, parentID
}

func parseOpenTracingHeaders(traceID, spanID string) (string, string) {
	traceID, ok := normalizeHexTraceIDRange(traceID, 1, 32)
	if !ok {
		return "", ""
	}
	return traceID, normalizeOptionalHexSpanID(spanID)
}

func parseXRayHeader(value string) (traceID, parentID string) {
	if strings.TrimSpace(value) == "" {
		return "", ""
	}

	parts := map[string]string{}
	for _, item := range strings.Split(value, ";") {
		key, val, ok := strings.Cut(strings.TrimSpace(item), "=")
		if ok && key != "" {
			parts[key] = val
		}
	}

	rootParts := strings.Split(parts["Root"], "-")
	if len(rootParts) != 3 {
		return "", ""
	}

	traceID, ok := normalizeHexTraceID(rootParts[1]+rootParts[2], 32)
	if !ok {
		return "", ""
	}
	return traceID, normalizeOptionalHexSpanID(parts["Parent"])
}

func decodeSW8Part(value string, maxDecodedLen int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range encodings {
		if len(value) > enc.EncodedLen(maxDecodedLen) {
			continue
		}

		if decoded, err := enc.DecodeString(value); err == nil && len(decoded) > 0 {
			if len(decoded) > maxDecodedLen || !utf8.Valid(decoded) {
				continue
			}
			return string(decoded)
		}
	}
	return ""
}

func normalizeDecimalTraceID(value string) string {
	id, normalized, ok := normalizeDatadogDecimalID(value)
	if !ok || id == 0 {
		return ""
	}
	return normalized
}

func normalizeDecimalSpanID(value string) string {
	_, normalized, ok := normalizeDatadogDecimalID(value)
	if !ok {
		return ""
	}
	return normalized
}

func normalizeDatadogDecimalID(value string) (uint64, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxDatadogDecimalIDLen {
		return 0, "", false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, "", false
		}
	}

	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, "", false
	}
	return id, value, true
}

func normalizeSW8SpanID(value string) string {
	if len(strings.TrimSpace(value)) > maxSW8SpanIDLen {
		return ""
	}
	return normalizeDecimalSpanID(value)
}

func normalizeHexTraceID(value string, lens ...int) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if !lengthIn(value, lens...) || isAll(value, '0') || !isHex(value) {
		return "", false
	}
	return value, true
}

func normalizeHexTraceIDRange(value string, minLen, maxLen int) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) < minLen || len(value) > maxLen || isAll(value, '0') || !isHex(value) {
		return "", false
	}
	return value, true
}

func normalizeHexSpanID(value string) (string, bool) {
	value = normalizeOptionalHexSpanID(value)
	if value == "" || isAll(value, '0') {
		return "", false
	}
	return value, true
}

func normalizeOptionalHexSpanID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 0 || len(value) > 16 || !isHex(value) {
		return ""
	}
	return value
}

func lengthIn(value string, lens ...int) bool {
	for _, ln := range lens {
		if len(value) == ln {
			return true
		}
	}
	return false
}

func isAll(value string, ch rune) bool {
	for _, c := range value {
		if c != ch {
			return false
		}
	}
	return value != ""
}

func isHex(value string) bool {
	for _, ch := range value {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}
