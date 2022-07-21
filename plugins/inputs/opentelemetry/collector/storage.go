// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collector

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

var (
	inputName = "opentelemetry"
	l         = logger.DefaultSLogger("otel-log")
	maxSend   = 100
	interval  = 10
)

// SpansStorage stores the spans.
type SpansStorage struct {
	AfterGather  *DKtrace.AfterGather
	RegexpString string
	CustomerTags []string
	GlobalTags   map[string]string
	traceMu      sync.Mutex
	rsm          DKtrace.DatakitTraces
	metricMu     sync.Mutex
	otelMetrics  []*OtelResourceMetric
	Count        int
	max          chan int
	stop         chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage() *SpansStorage {
	l = logger.SLogger(inputName)
	return &SpansStorage{
		AfterGather: DKtrace.NewAfterGather(),
		rsm:         make(DKtrace.DatakitTraces, 0),
		otelMetrics: make([]*OtelResourceMetric, 0),
		max:         make(chan int, 1),
		stop:        make(chan struct{}, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (s *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := s.mkDKTrace(rss)
	s.traceMu.Lock()
	l.Debugf("mktrace %d otel span and add %d span to storage", len(rss), len(traces))
	s.rsm = append(s.rsm, traces...)
	s.traceMu.Unlock()
	s.Count += len(traces)
	if s.Count >= maxSend {
		s.max <- 0
	}
}

func (s *SpansStorage) AddMetric(rss []*OtelResourceMetric) {
	s.metricMu.Lock()
	l.Debugf("AddMetric add %d metric to storage", len(rss))
	s.otelMetrics = append(s.otelMetrics, rss...)
	s.metricMu.Unlock()
	s.Count += len(rss)
	if s.Count >= maxSend {
		s.max <- 0
	}
}

// GetDKTrace  returns the stored resource spans.
func (s *SpansStorage) GetDKTrace() DKtrace.DatakitTraces {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	rss := make(DKtrace.DatakitTraces, 0, len(s.rsm))
	rss = append(rss, s.rsm...)
	s.rsm = s.rsm[:0]

	return rss
}

func (s *SpansStorage) GetDKMetric() []*OtelResourceMetric {
	s.metricMu.Lock()
	defer s.metricMu.Unlock()
	rss := make([]*OtelResourceMetric, 0, len(s.rsm))
	rss = append(rss, s.otelMetrics...)
	s.otelMetrics = s.otelMetrics[:0]
	return rss
}

func (s *SpansStorage) getCount() int {
	return s.Count
}

func (s *SpansStorage) Run() {
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

// feedAll : trace -> io.trace  |  metric -> io.
func (s *SpansStorage) feedAll() {
	traces := s.GetDKTrace()
	s.AfterGather.Run(inputName, traces, false)
	l.Debugf("send %d trace to io.trace", len(traces))

	if metrics := s.GetDKMetric(); len(metrics) > 0 {
		pts := makePoints(metrics)
		err := dkio.Feed(inputName, datakit.Metric, pts, &dkio.Option{HighFreq: true})
		if err != nil {
			l.Errorf("feed to io error=%v", err)
		}
	}
	s.Count = 0
}
