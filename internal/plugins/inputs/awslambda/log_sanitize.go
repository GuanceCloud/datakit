// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"regexp"
	"strings"
)

const redactedValue = "******"

var (
	sensitiveLogKeyPattern = regexp.MustCompile(
		`(?i)(^|[_\-.])(api[_\-.]?key|access[_\-.]?key|secret|token|password)([_\-.]|$)`,
	)
	keyValueSecretPattern = regexp.MustCompile(
		`(?i)\b([A-Za-z0-9_.-]*(?:api[_-]?key|access[_-]?key|secret|token|password)[A-Za-z0-9_.-]*|ENV_DATAWAY|DD_API_KEY)\s*=\s*([^\s,;)\]]+)`,
	)
	jsonSecretPattern = regexp.MustCompile(
		`(?i)("([A-Za-z0-9_.-]*(?:api[_-]?key|access[_-]?key|secret|token|password)[A-Za-z0-9_.-]*|ENV_DATAWAY|DD_API_KEY)"\s*:\s*")([^"]*)(")`,
	)
	escapedJSONSecretPattern = regexp.MustCompile(
		`(?i)(\\"([A-Za-z0-9_.-]*(?:api[_-]?key|access[_-]?key|secret|token|password)[A-Za-z0-9_.-]*|ENV_DATAWAY|DD_API_KEY)\\"\s*:\s*\\")([^\\"]*)(\\")`,
	)
	urlSecretPattern = regexp.MustCompile(
		`(?i)([?&](?:token|api[_-]?key|apikey|access[_-]?token|password|secret)=)[^&\s"\\]+`,
	)
)

func sanitizeLogFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return fields
	}

	sanitized := make(map[string]any, len(fields))
	for k, v := range fields {
		if isSensitiveLogKey(k) {
			sanitized[k] = redactedValue
			continue
		}
		if s, ok := v.(string); ok {
			sanitized[k] = sanitizeLogMessage(s)
			continue
		}
		sanitized[k] = v
	}
	return sanitized
}

func sanitizeLogMessage(message string) string {
	if message == "" {
		return message
	}

	message = keyValueSecretPattern.ReplaceAllString(message, `${1}=`+redactedValue)
	message = jsonSecretPattern.ReplaceAllString(message, `${1}`+redactedValue+`${4}`)
	message = escapedJSONSecretPattern.ReplaceAllString(message, `${1}`+redactedValue+`${4}`)
	message = urlSecretPattern.ReplaceAllString(message, `${1}`+redactedValue)
	return message
}

func isSensitiveLogKey(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	if strings.EqualFold(key, "ENV_DATAWAY") || strings.EqualFold(key, "DD_API_KEY") {
		return true
	}
	return sensitiveLogKeyPattern.MatchString(key)
}
