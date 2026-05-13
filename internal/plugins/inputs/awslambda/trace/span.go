// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdatrace

type Span struct {
	TraceID  uint64             `json:"trace_id" msgpack:"trace_id"`
	SpanID   uint64             `json:"span_id" msgpack:"span_id"`
	ParentID uint64             `json:"parent_id" msgpack:"parent_id"`
	Name     string             `json:"name" msgpack:"name"`
	Resource string             `json:"resource" msgpack:"resource"`
	Service  string             `json:"service" msgpack:"service"`
	Type     string             `json:"type" msgpack:"type"`
	Start    int64              `json:"start" msgpack:"start"`
	Duration int64              `json:"duration" msgpack:"duration"`
	Error    int32              `json:"error" msgpack:"error"`
	Meta     map[string]string  `json:"meta" msgpack:"meta"`
	Metrics  map[string]float64 `json:"metrics" msgpack:"metrics"`
}
