// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import "sync"

//go:generate msgp -file=span.pb.go -o span_gen.go -io=false
//go:generate msgp -o trace_gen.go -io=false

type DDTrace []*DDSpan

type DDTraces []DDTrace

var ddtracePool = &sync.Pool{
	New: func() interface{} {
		return DDTraces{}
	},
}

func (t DDTraces) reset() {
	for _, trace := range t {
		for _, span := range trace {
			span.Service = ""
			span.Name = ""
			span.Resource = ""
			span.TraceID = 0
			span.SpanID = 0
			span.ParentID = 0
			span.Start = 0
			span.Duration = 0
			span.Error = 0
			for s := range span.Meta {
				span.Meta[s] = ""
			}
			for s := range span.Metrics {
				span.Metrics[s] = 0
			}
			span.Type = ""
		}
		// trace = trace[:0]
	}
}
