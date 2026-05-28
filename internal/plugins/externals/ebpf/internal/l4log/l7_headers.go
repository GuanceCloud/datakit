//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	maxL7RecordedHeaders = 32
	maxL7HeaderValueLen  = 512
)

var defaultL7RecordedHeaderAllowlist = map[string]struct{}{
	"host":                          {},
	":authority":                    {},
	"forwarded":                     {},
	"x-forwarded-for":               {},
	"x-forwarded-host":              {},
	"x-forwarded-port":              {},
	"x-forwarded-proto":             {},
	"x-original-forwarded-for":      {},
	"x-real-ip":                     {},
	"x-request-id":                  {},
	"x-correlation-id":              {},
	"x-correlationid":               {},
	"x-trace-id":                    {},
	"traceparent":                   {},
	"tracestate":                    {},
	"x-datadog-trace-id":            {},
	"x-datadog-parent-id":           {},
	"x-datadog-span-id":             {},
	"x-datadog-sampling-priority":   {},
	"x-datadog-tags":                {},
	"x-datadog-origin":              {},
	"b3":                            {},
	"x-b3-traceid":                  {},
	"x-b3-spanid":                   {},
	"x-b3-parentspanid":             {},
	"x-b3-sampled":                  {},
	"x-b3-flags":                    {},
	"sw8":                           {},
	"uber-trace-id":                 {},
	"ot-tracer-traceid":             {},
	"ot-tracer-spanid":              {},
	"x-ot-span-context":             {},
	"x-amzn-trace-id":               {},
	"x-cloud-trace-context":         {},
	"user-agent":                    {},
	"accept":                        {},
	"accept-encoding":               {},
	"content-type":                  {},
	"content-length":                {},
	"te":                            {},
	"grpc-accept-encoding":          {},
	"grpc-encoding":                 {},
	"grpc-message":                  {},
	"grpc-status":                   {},
	"x-envoy-attempt-count":         {},
	"x-envoy-external-address":      {},
	"x-envoy-original-path":         {},
	"x-envoy-upstream-service-time": {},
}

var l7RecordedHeaderAllowlist = cloneL7LogHeaderAllowlist(defaultL7RecordedHeaderAllowlist)

var l7SensitiveHeaders = map[string]struct{}{
	"authorization":         {},
	"cookie":                {},
	"proxy-authorization":   {},
	"set-cookie":            {},
	"x-api-key":             {},
	"x-auth-token":          {},
	"x-csrf-token":          {},
	"x-goog-api-key":        {},
	"x-xsrf-token":          {},
	"x-amz-security-token":  {},
	"x-hub-signature":       {},
	"x-hub-signature-256":   {},
	"x-signature":           {},
	"x-webhook-signature":   {},
	"x-slack-signature":     {},
	"x-gitlab-token":        {},
	"x-shopify-hmac-sha256": {},
}

func configureL7LogHeaders(headers []string) {
	if len(headers) == 0 {
		l7RecordedHeaderAllowlist = cloneL7LogHeaderAllowlist(defaultL7RecordedHeaderAllowlist)
		return
	}

	configured := make(map[string]struct{}, len(headers))
	haveExplicit := false
	for _, header := range headers {
		for _, part := range strings.Split(header, ",") {
			key := strings.ToLower(strings.TrimSpace(part))
			if key == "" || key == "default" {
				continue
			}
			haveExplicit = true

			switch key {
			case "none", "off", "disable", "disabled":
				l7RecordedHeaderAllowlist = map[string]struct{}{}
				return
			}

			key, ok := configuredL7LogHeaderKey(key)
			if ok {
				configured[key] = struct{}{}
			}
		}
	}

	if !haveExplicit {
		l7RecordedHeaderAllowlist = cloneL7LogHeaderAllowlist(defaultL7RecordedHeaderAllowlist)
		return
	}
	l7RecordedHeaderAllowlist = configured
}

func recordL7LogHeaderBytes(headers map[string]string, name, value []byte) map[string]string {
	key, ok := l7LogHeaderKeyBytes(name)
	if !ok {
		return headers
	}
	return recordL7LogHeaderValue(headers, key, string(value))
}

func recordL7LogHeaderString(headers map[string]string, name, value string) map[string]string {
	key, ok := l7LogHeaderKeyString(name)
	if !ok {
		return headers
	}
	return recordL7LogHeaderValue(headers, key, value)
}

func mergeL7LogHeaders(dst, src map[string]string) map[string]string {
	if len(src) == 0 {
		return dst
	}

	for key, value := range src {
		dst = recordL7LogHeaderValue(dst, key, value)
	}
	return dst
}

func l7LogHeadersSummary(req, resp map[string]string) (count, size int) {
	for key, value := range req {
		count++
		size += len(key) + len(value)
	}
	for key, value := range resp {
		count++
		size += len(key) + len(value)
	}
	return count, size
}

func l7HeaderValue(headers map[string]string, key string) string {
	if len(headers) == 0 {
		return ""
	}
	return headers[key]
}

func l7ContentLength(headers map[string]string) *int64 {
	value := l7HeaderValue(headers, "content-length")
	if value == "" {
		return nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || n < 0 {
		return nil
	}
	return &n
}

func recordL7LogHeaderValue(headers map[string]string, key, value string) map[string]string {
	value = normalizeL7LogHeaderValue(value)
	if value == "" {
		return headers
	}

	if _, exists := headers[key]; exists {
		return headers
	}
	if len(headers) >= maxL7RecordedHeaders {
		return headers
	}
	if headers == nil {
		headers = make(map[string]string, 8)
	}
	headers[key] = value
	return headers
}

func cloneL7LogHeaderAllowlist(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{}, len(src))
	for key := range src {
		dst[key] = struct{}{}
	}
	return dst
}

func l7LogHeaderKeyBytes(name []byte) (string, bool) {
	name = bytes.TrimSpace(name)
	if len(name) == 0 {
		return "", false
	}
	return l7LogHeaderKeyString(string(name))
}

func l7LogHeaderKeyString(name string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return "", false
	}
	if key == ":authority" {
		key = "host"
	}

	if _, ok := l7SensitiveHeaders[key]; ok {
		return "", false
	}
	if _, ok := l7RecordedHeaderAllowlist[key]; !ok {
		return "", false
	}
	return key, true
}

func configuredL7LogHeaderKey(key string) (string, bool) {
	if key == ":authority" {
		key = "host"
	}
	if _, ok := l7SensitiveHeaders[key]; ok {
		return "", false
	}
	if !validL7LogHeaderName(key) {
		return "", false
	}
	return key, true
}

func validL7LogHeaderName(key string) bool {
	if key == "" {
		return false
	}
	if strings.HasPrefix(key, ":") {
		return key == ":authority"
	}
	for _, c := range key {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", c):
		default:
			return false
		}
	}
	return true
}

func normalizeL7LogHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	value = strings.Map(func(r rune) rune {
		switch r {
		case '\r', '\n':
			return -1
		default:
			return r
		}
	}, value)
	if len(value) > maxL7HeaderValueLen {
		value = value[:maxL7HeaderValueLen]
		value = strings.ToValidUTF8(value, "")
	}
	return value
}
