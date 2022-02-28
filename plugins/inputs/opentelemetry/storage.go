// Package opentelemetry storage

package opentelemetry

import (
	"sync"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// SpansStorage stores the spans.
type SpansStorage struct {
	traceMu     sync.Mutex
	rsm         []DKtrace.DatakitTrace
	metricMu    sync.Mutex
	otelMetrics []*otelResourceMetric
	Count       int
	max         chan int
	stop        chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage() *SpansStorage {
	return &SpansStorage{
		rsm:         make([]DKtrace.DatakitTrace, 0),
		otelMetrics: make([]*otelResourceMetric, 0),
		max:         make(chan int, 1),
		stop:        make(chan struct{}, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (s *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := mkDKTrace(rss)
	s.traceMu.Lock()
	s.rsm = append(s.rsm, traces...)
	s.traceMu.Unlock()
	s.Count += len(traces)
	if s.Count >= maxSend {
		s.max <- 0
	}
}

func (s *SpansStorage) AddMetric(rss []*otelResourceMetric) {
	s.metricMu.Lock()
	s.otelMetrics = append(s.otelMetrics, rss...)
	s.metricMu.Unlock()
	s.Count += len(rss)
	if s.Count >= maxSend {
		s.max <- 0
	}
}

// GetResourceSpans returns the stored resource spans.
func (s *SpansStorage) getDKTrace() []DKtrace.DatakitTrace {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	rss := make([]DKtrace.DatakitTrace, 0, len(s.rsm))
	rss = append(rss, s.rsm...)
	s.rsm = s.rsm[:0]
	return rss
}

func (s *SpansStorage) getDKMetric() []*otelResourceMetric {
	s.metricMu.Lock()
	defer s.metricMu.Unlock()
	rss := make([]*otelResourceMetric, 0, len(s.rsm))
	rss = append(rss, s.otelMetrics...)
	s.otelMetrics = s.otelMetrics[:0]
	return rss
}

func (s *SpansStorage) getCount() int {
	return s.Count
}

func (s *SpansStorage) run() {
	for {
		select {
		case <-s.max:
			s.feedAll()
		case <-time.After(time.Duration(interval) * time.Second):
			if s.getCount() > 0 {
				s.feedAll()
			}
		case <-s.stop:
			l.Infof("spanStorage stop")
			close(s.stop)
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
	metrics := s.getDKMetric()
	if len(metrics) > 0 {
		pts := makePoints(metrics)
		err := dkio.Feed(inputName, datakit.Metric, pts, &dkio.Option{HighFreq: true})
		// err := inputs.FeedMeasurement(inputName, datakit.Metric, metrics, &dkio.Option{HighFreq: true})
		if err != nil {
			l.Errorf("feed to io error=%v", err)
		}
	}
	s.Count = 0
}
