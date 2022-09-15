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
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

var (
	inputName = "opentelemetry"
	log       = logger.DefaultSLogger("otel-log")
	maxSend   = 100
	interval  = 10
)

// SpansStorage stores the spans.
type SpansStorage struct {
	AfterGather  *itrace.AfterGather
	RegexpString string
	CustomerTags []string
	GlobalTags   map[string]string
	traceMu      sync.Mutex
	rsm          itrace.DatakitTraces
	metricMu     sync.Mutex
	otelMetrics  []*OtelResourceMetric
	Count        int
	max          chan int
	stop         chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage(afaterGather *itrace.AfterGather) *SpansStorage {
	log = logger.SLogger(inputName)

	return &SpansStorage{
		AfterGather: afaterGather,
		rsm:         make(itrace.DatakitTraces, 0),
		otelMetrics: make([]*OtelResourceMetric, 0),
		max:         make(chan int, 1),
		stop:        make(chan struct{}, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (ss *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := ss.mkDKTrace(rss)
	ss.traceMu.Lock()
	log.Debugf("mktrace %d otel span and add %d span to storage", len(rss), len(traces))
	ss.rsm = append(ss.rsm, traces...)
	ss.traceMu.Unlock()
	ss.Count += len(traces)
	if ss.Count >= maxSend {
		ss.max <- 0
	}
}

func (ss *SpansStorage) AddMetric(rss []*OtelResourceMetric) {
	ss.metricMu.Lock()
	log.Debugf("AddMetric add %d metric to storage", len(rss))
	ss.otelMetrics = append(ss.otelMetrics, rss...)
	ss.metricMu.Unlock()
	ss.Count += len(rss)
	if ss.Count >= maxSend {
		ss.max <- 0
	}
}

// GetDKTrace  returns the stored resource spans.
func (ss *SpansStorage) GetDKTrace() itrace.DatakitTraces {
	ss.traceMu.Lock()
	defer ss.traceMu.Unlock()
	rss := make(itrace.DatakitTraces, 0, len(ss.rsm))
	rss = append(rss, ss.rsm...)
	ss.rsm = ss.rsm[:0]

	return rss
}

func (ss *SpansStorage) GetDKMetric() []*OtelResourceMetric {
	ss.metricMu.Lock()
	defer ss.metricMu.Unlock()
	rss := make([]*OtelResourceMetric, 0, len(ss.rsm))
	rss = append(rss, ss.otelMetrics...)
	ss.otelMetrics = ss.otelMetrics[:0]

	return rss
}

func (ss *SpansStorage) getCount() int {
	return ss.Count
}

func (ss *SpansStorage) Run() {
	for {
		select {
		case <-ss.max:
			ss.feedAll()
		case <-time.After(time.Duration(interval) * time.Second):
			if ss.getCount() > 0 {
				ss.feedAll()
			}
		case <-ss.stop:
			log.Infof("spanStorage stop")
			close(ss.stop)

			return
		}
	}
}

func (s *SpansStorage) Stop() {
	s.stop <- struct{}{}
}

// feedAll : trace -> io.trace  |  metric -> io.
func (ss *SpansStorage) feedAll() {
	traces := ss.GetDKTrace()
	ss.AfterGather.Run(inputName, traces, false)
	log.Debugf("send %d trace to io.trace", len(traces))

	if metrics := ss.GetDKMetric(); len(metrics) > 0 {
		pts := makePoints(metrics)
		err := dkio.Feed(inputName, datakit.Metric, pts, &dkio.Option{HighFreq: true})
		if err != nil {
			log.Errorf("feed to io error=%v", err)
		}
	}
	ss.Count = 0
}
