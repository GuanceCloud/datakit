// Package opentelemetry storage

package opentelemetry

import (
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// SpansStorage stores the spans.
type SpansStorage struct {
	rsm         []DKtrace.DatakitTrace
	otelMetrics []*otelResourceMetric
	Count       int
	max         chan int
	stop        chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage() SpansStorage {
	return SpansStorage{
		rsm:         make([]DKtrace.DatakitTrace, 0),
		otelMetrics: make([]*otelResourceMetric, 0),
		max:         make(chan int, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (s *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := mkDKTrace(rss)
	s.rsm = append(s.rsm, traces...)
	s.Count += len(traces)
	if s.Count >= maxSend {
		s.max <- 0
	}
}

func (s *SpansStorage) AddMetric(rss []*otelResourceMetric) {
	s.otelMetrics = append(s.otelMetrics, rss...)
	s.Count += len(rss)
	if s.Count >= maxSend {
		s.max <- 0
	}
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

func (s *SpansStorage) getDKMetric() []*otelResourceMetric {
	rss := make([]*otelResourceMetric, 0, len(s.rsm))
	rss = append(rss, s.otelMetrics...)
	return rss
}

func (s *SpansStorage) getMeasurementMetrics() []inputs.Measurement {
	ms := make([]inputs.Measurement, 0)
	dkMs := otelMetricToDkMetric(s.otelMetrics)
	for _, metric := range dkMs {
		ms = append(ms, metric)
	}
	return ms
}

func (s *SpansStorage) getCount() int {
	return s.Count
}

func (s *SpansStorage) run() {
	for {
		select {
		case <-s.max:
			s.feedAll()
			s.reset()
		case <-time.After(time.Duration(interval) * time.Second):
			if s.getCount() > 0 {
				s.feedAll()
				s.reset()
			}
		case <-s.stop:
			return
		}
	}
}

// feedAll : trace -> io.trace  |  metric -> io
func (s *SpansStorage) feedAll() {
	traces := s.getDKTrace()
	for _, trace := range traces {
		afterGather.Run(inputName, trace, false)
	}
	metrics := s.getMeasurementMetrics()
	if len(metrics) > 0 {
		err := inputs.FeedMeasurement(inputName, datakit.Metric, metrics, &dkio.Option{HighFreq: true})
		if err != nil {
			l.Errorf("feed to io error=%v", err)

		}
	}
}

func (s *SpansStorage) reset() {
	// 归零
	s.Count = 0
	s.rsm = s.rsm[:0]
	s.otelMetrics = s.otelMetrics[:0]
}
