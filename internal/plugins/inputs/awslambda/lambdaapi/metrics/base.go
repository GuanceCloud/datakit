// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package metrics defines constants used for tracking various metrics across different systems.
// These metrics include memory usage, runtime duration, response latency, and more, allowing for
// comprehensive monitoring and analysis of system performance.
package metrics

const (
	MaxMemoryUsedMetric       = "max_memory_used_mb"
	MemorySizeMetric          = "memory_size_mb"
	RuntimeDurationMetric     = "runtime_duration_ms"
	BilledDurationMetric      = "billed_duration_ms"
	DurationMetric            = "duration_ms"
	PostRuntimeDurationMetric = "post_runtime_duration"
	InitDurationMetric        = "init_duration_ms"
	ResponseLatencyMetric     = "response_latency"
	ResponseDurationMetric    = "response_duration_ms"
	ProducedBytesMetric       = "produced_bytes"
	OutOfMemoryMetric         = "out_of_memory"
	TimeoutsMetric            = "timeouts"
	ErrorsMetric              = "errors"
	InvocationsMetric         = "invocations"
)
