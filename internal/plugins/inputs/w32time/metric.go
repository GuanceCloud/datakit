// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package w32time

import "math"

const (
	fieldServiceRunning       = "service_running"
	fieldComputedTimeOffset   = "computed_time_offset"
	fieldNTPRoundtripDelay    = "ntp_roundtrip_delay"
	fieldNTPClientSourceCount = "ntp_client_source_count"

	microsecondsPerSecond = 1_000_000.0
)

func normalizeCounterValues(raw map[string]float64) map[string]interface{} {
	fields := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		switch key {
		case fieldComputedTimeOffset, fieldNTPRoundtripDelay:
			fields[key] = value / microsecondsPerSecond
		case fieldNTPClientSourceCount:
			fields[key] = int64(math.Round(value))
		default:
			fields[key] = value
		}
	}

	return fields
}
