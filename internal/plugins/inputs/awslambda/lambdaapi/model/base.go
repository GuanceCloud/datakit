// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package model define some of the models used by the extension api and telemetry api.
package model

import (
	"fmt"
	"strings"
	"time"
	"unsafe"
)

// EventType represents the type of events received from /event/next.
type EventType string

const (
	// Invoke is a lambda invoke.
	Invoke EventType = "INVOKE"

	// Shutdown is a shutdown event for the environment.
	Shutdown EventType = "SHUTDOWN"
)

type EventShutdown string

const (
	EventShutdownSpindown EventShutdown = "spindown"
	EventShutdownTimeout  EventShutdown = "timeout"
	EventShutdownFailure  EventShutdown = "failure"
)

type InitType string

const (
	InitTypeOnDemand               InitType = "on-demand"
	InitTypeProvisionedConcurrency InitType = "provisioned-concurrency"
	InitTypeSnapStart              InitType = "snap-start"
)

type InitPhase string

const (
	InitPhaseInit   InitPhase = "init"
	InitPhaseInvoke InitPhase = "invoke"
)

type TracingType string

const (
	TracingTypeXAmznTraceID TracingType = "X-Amzn-Trace-Id"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusError   Status = "error"
	StatusTimeout Status = "timeout"
)

type SubscriptionEventType string

const (
	SubscriptionEventPlatform  SubscriptionEventType = "platform"
	SubscriptionEventFunction  SubscriptionEventType = "function"
	SubscriptionEventExtension SubscriptionEventType = "extension"
)

// TraceInfo used to store tracking information.
type TraceInfo struct {
	Root    string
	Parent  string
	Sampled string
}

// Tracing is part of the response for /event/next.
type Tracing struct {
	Type  TracingType `json:"type"`
	Value TraceInfo   `json:"value"`
}

type TracingContext struct {
	SpanID string `json:"spanId,omitempty"`
	Tracing
}

type Span struct {
	DurationMs float64   `json:"durationMs"`
	Name       string    `json:"name"`
	Start      time.Time `json:"start"`
}

type BaseMetrics struct {
	DurationMs float64 `json:"durationMs"`
}

// InitReportMetrics holds initialization report metrics.
type InitReportMetrics struct {
	BaseMetrics
}

type RestoreReportMetrics struct {
	BaseMetrics
}

// RuntimeDoneMetrics holds runtime done metrics.
type RuntimeDoneMetrics struct {
	BaseMetrics
	ProducedBytes uint64 `json:"producedBytes,omitempty"`
}

// ReportMetrics holds metrics for report records.
type ReportMetrics struct {
	BaseMetrics
	BilledDurationMs        uint64   `json:"billedDurationMs"`
	BilledRestoreDurationMs *uint64  `json:"billedRestoreDurationMs,omitempty"`
	InitDurationMs          *float64 `json:"initDurationMs,omitempty"`
	MaxMemoryUsedMB         uint64   `json:"maxMemoryUsedMB"`
	MemorySizeMB            uint64   `json:"memorySizeMB"`
	RestoreDurationMs       *float64 `json:"restoreDurationMs,omitempty"`
}

// parseTraceInfo convert formatted strings into TraceInfo.
func (t *TraceInfo) parseTraceInfo(s string) error {
	pairs := strings.Split(s, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return fmt.Errorf("invalid pair: %s", pair)
		}
		key, value := kv[0], kv[1]
		switch key {
		case "Root":
			t.Root = value
		case "Parent":
			t.Parent = value
		case "Sampled":
			t.Sampled = value
		}
	}
	return nil
}

func (t *TraceInfo) UnmarshalJSON(data []byte) error {
	p := data[1 : len(data)-1]
	s := *(*string)(unsafe.Pointer(&p)) //nolint:gosec
	return t.parseTraceInfo(s)
}
