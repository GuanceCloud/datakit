// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

const (
	SpanStatusInfo  = "info"
	SpanStatusOK    = "ok"
	SpanStatusError = "error"
)

// TraceStatusFromCode maps OTLP trace status code to a compact status string.
func TraceStatusFromCode(code int32) string {
	switch code {
	case 0, 1:
		return SpanStatusOK
	case 2:
		return SpanStatusError
	default:
		return SpanStatusInfo
	}
}

// LogStatusFromSeverityNumber maps OTLP log severity numbers to a status string.
func LogStatusFromSeverityNumber(severityNumber int32, fallback string) string {
	switch {
	case severityNumber >= 1 && severityNumber <= 4:
		return "trace"
	case severityNumber >= 5 && severityNumber <= 8:
		return "debug"
	case severityNumber >= 9 && severityNumber <= 12:
		return "info"
	case severityNumber >= 13 && severityNumber <= 16:
		return "warn"
	case severityNumber >= 17 && severityNumber <= 20:
		return "error"
	case severityNumber >= 21 && severityNumber <= 24:
		return "fatal"
	case severityNumber == 0:
		return "unknown"
	default:
		return fallback
	}
}
