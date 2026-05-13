// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"strings"
	"testing"
)

func TestSanitizeLogMessage(t *testing.T) {
	input := strings.Join([]string{
		"AWS_ACCESS_KEY_ID=AKIA_TEST",
		"AWS_SECRET_ACCESS_KEY=secret-value",
		"AWS_SESSION_TOKEN=session-token",
		"DD_API_KEY=dd-api-key",
		"ENV_DATAWAY=http://openway.guance.com?token=tkn_secret",
		`{"DD_API_KEY":"json-api-key","normal":"value"}`,
		`{\"AWS_SESSION_TOKEN\":\"escaped-token\",\"normal\":\"value\"}`,
	}, "\n")

	got := sanitizeLogMessage(input)

	for _, leaked := range []string{
		"AKIA_TEST",
		"secret-value",
		"session-token",
		"dd-api-key",
		"tkn_secret",
		"json-api-key",
		"escaped-token",
	} {
		if strings.Contains(got, leaked) {
			t.Fatalf("sanitized message still contains %q: %s", leaked, got)
		}
	}

	for _, kept := range []string{
		"AWS_ACCESS_KEY_ID=" + redactedValue,
		"AWS_SECRET_ACCESS_KEY=" + redactedValue,
		"AWS_SESSION_TOKEN=" + redactedValue,
		"DD_API_KEY=" + redactedValue,
		`"normal":"value"`,
		`\"normal\":\"value\"`,
	} {
		if !strings.Contains(got, kept) {
			t.Fatalf("sanitized message does not contain %q: %s", kept, got)
		}
	}
}

func TestSanitizeLogFields(t *testing.T) {
	fields := map[string]any{
		"message":      "ENV_DATAWAY=http://openway.guance.com?token=tkn_secret normal=value",
		"DD_API_KEY":   "dd-api-key",
		"normal_field": "normal-value",
		"count":        3,
	}

	got := sanitizeLogFields(fields)

	if got["DD_API_KEY"] != redactedValue {
		t.Fatalf("expected DD_API_KEY to be redacted, got %v", got["DD_API_KEY"])
	}
	if got["normal_field"] != "normal-value" {
		t.Fatalf("expected normal field to be preserved, got %v", got["normal_field"])
	}
	if got["count"] != 3 {
		t.Fatalf("expected non-string field to be preserved, got %v", got["count"])
	}
	message, _ := got["message"].(string)
	if strings.Contains(message, "tkn_secret") {
		t.Fatalf("message token was not redacted: %s", message)
	}
}
