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

const (
	maxPooledTraceCount     = 512
	maxPooledTraceSpanSlots = 4096
	maxPooledTotalSpanSlots = 16384
	maxPooledMapEntries     = 64
)

var ddtracePool = &sync.Pool{
	New: func() interface{} {
		return DDTraces{}
	},
}

func (t DDTraces) shouldKeepInPool() bool {
	if cap(t) > maxPooledTraceCount {
		return false
	}

	totalSpanSlots := 0
	for i := range t {
		if cap(t[i]) > maxPooledTraceSpanSlots {
			return false
		}

		totalSpanSlots += cap(t[i])
		if totalSpanSlots > maxPooledTotalSpanSlots {
			return false
		}
	}

	return true
}

func (t *DDTraces) reset(keepInPool bool) {
	if t == nil {
		return
	}

	for i := range *t {
		trace := (*t)[i]
		for j := range trace {
			if trace[j] == nil {
				continue
			}

			resetDDSpan(trace[j], keepInPool)
			if !keepInPool {
				trace[j] = nil
			}
		}

		if keepInPool {
			(*t)[i] = trace[:0]
		} else {
			(*t)[i] = nil
		}
	}

	if keepInPool {
		*t = (*t)[:0]
	} else {
		*t = nil
	}
}

func resetDDSpan(span *DDSpan, keepInPool bool) {
	if span == nil {
		return
	}

	if !keepInPool {
		*span = DDSpan{}
		return
	}

	span.Service = ""
	span.Name = ""
	span.Resource = ""
	span.TraceID = 0
	span.SpanID = 0
	span.ParentID = 0
	span.Start = 0
	span.Duration = 0
	span.Error = 0
	span.Type = ""

	if len(span.Meta) > maxPooledMapEntries {
		span.Meta = nil
	} else {
		for k := range span.Meta {
			delete(span.Meta, k)
		}
	}

	if len(span.Metrics) > maxPooledMapEntries {
		span.Metrics = nil
	} else {
		for k := range span.Metrics {
			delete(span.Metrics, k)
		}
	}
}

// wrapMessage DataKit issue:#2879
// 其他字段都已经提取，除了这两个 map:
// meta 是提取到一级字段之后剩余的字段
// metrics 是没有提取过一级字段的。
type wrapMessage struct {
	Meta    map[string]string  `json:"meta"`
	Metrics map[string]float64 `json:"metrics"`
}
