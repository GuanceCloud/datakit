package util

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

var nonKeyChar = regexp.MustCompile(`[^A-Za-z0-9_]+`)
var repeatedUnderscore = regexp.MustCompile(`_+`)

func SanitizeKey(input string, fallback string) string {
	cleaned := strings.TrimSpace(input)
	cleaned = camelToSnake(cleaned)
	cleaned = nonKeyChar.ReplaceAllString(cleaned, "_")
	cleaned = repeatedUnderscore.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "_")
	cleaned = strings.ToLower(cleaned)
	if cleaned == "" {
		cleaned = fallback
	}
	if cleaned != "" && cleaned[0] >= '0' && cleaned[0] <= '9' {
		return "k_" + cleaned
	}
	return cleaned
}

func SanitizeTags(tags map[string]string) map[string]string {
	out := make(map[string]string, len(tags))
	for key, value := range tags {
		out[SanitizeKey(key, "tag")] = value
	}
	return out
}

func SanitizeFields(fields map[string]any) map[string]any {
	out := make(map[string]any, len(fields))
	for key, value := range fields {
		if value == nil {
			continue
		}
		out[SanitizeKey(key, "field")] = value
	}
	return out
}

func Truncate(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	suffix := "...[truncated]"
	if maxLen <= len(suffix) {
		return value[:maxLen]
	}
	return value[:maxLen-len(suffix)] + suffix
}

func JSONString(value any, maxLen int) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return Truncate(`"`+err.Error()+`"`, maxLen)
	}
	return Truncate(string(bytes), maxLen)
}

func camelToSnake(input string) string {
	var out strings.Builder
	var prev rune
	for i, current := range input {
		if i > 0 && unicode.IsUpper(current) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			out.WriteRune('_')
		}
		out.WriteRune(current)
		prev = current
	}
	return out.String()
}
