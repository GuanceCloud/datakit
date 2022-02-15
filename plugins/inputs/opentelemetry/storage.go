// Package opentelemetry storage

package opentelemetry

import (
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// SpansStorage stores the spans
type SpansStorage struct {
	rsm       []DKtrace.DatakitTrace
	dkMetric  []*otelResourceMetric
	spanCount int
	max       chan int
	stop      chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage() SpansStorage {
	return SpansStorage{
		rsm:      make([]DKtrace.DatakitTrace, 0),
		dkMetric: make([]*otelResourceMetric, 0),
		max:      make(chan int, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (s *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := mkDKTrace(rss)
	s.rsm = append(s.rsm, traces...)
	s.spanCount += len(traces)
	if s.spanCount >= maxSend {
		s.max <- 0
	}
}

func (s *SpansStorage) AddMetric(rss []*otelResourceMetric) {
	// todo
}

func setTag(tags map[string]string, attr []*commonpb.KeyValue) map[string]string {
	for _, kv := range attr {
		for _, tag := range customTags {
			if kv.Key == tag {
				key := replace(kv.Key)
				tags[key] = kv.GetValue().String()
			}
		}
	}
	for k, v := range globalTags {
		tags[k] = v
	}
	return tags
}

// GetResourceSpans returns the stored resource spans.
func (s *SpansStorage) getDKTrace() []DKtrace.DatakitTrace {
	rss := make([]DKtrace.DatakitTrace, 0, len(s.rsm))
	rss = append(rss, s.rsm...)
	return rss
}

func (s *SpansStorage) getCount() int {
	return s.spanCount
}

func (s *SpansStorage) run() {
	// 定时发送 或者长度超过100
	for {
		select {
		case <-s.max:
			traces := s.getDKTrace()
			for _, trace := range traces {
				afterGather.Run(inputName, trace, false)
			}
			s.reset()
		case <-time.After(time.Duration(interval) * time.Second):
			if s.getCount() > 0 {
				traces := s.getDKTrace()
				for _, trace := range traces {
					afterGather.Run(inputName, trace, false)
				}
				s.reset()
			}
		case <-s.stop:
			return
		}
	}
}

func (s *SpansStorage) reset() {
	// 归零
	s.spanCount = 0
	s.rsm = s.rsm[:0]
}
