// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	cliMetrics "github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	profileMetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
)

const (
	profileObsUnknown       = "unknown"
	profileObsStatusCached  = "cached"
	profileObsStatusDropped = "dropped"
	profileObsStatusFailed  = "failed"
	profileObsStatusOK      = "ok"
	profileObsStatusRetry   = "retry"

	profileObsReasonMetadata  = "metadata"
	profileObsReasonMultipart = "multipart"
)

var (
	profileReceivedTotal,
	profileReceivedBytesTotal,
	profileParsedTotal,
	profileParsedBytesTotal,
	profileParseErrorsTotal,
	profileForwardTotal,
	profileForwardBytesTotal *prometheus.CounterVec

	profileForwardLatency *prometheus.SummaryVec

	profileObsSummary = newProfileObservabilitySummary()
)

type profileMetricLabels struct {
	language string
	format   string
	profiler string
}

type profileSummaryKey struct {
	language string
	format   string
	profiler string
}

type profileSummaryBucket struct {
	parsedCount    int64
	parsedBytes    int64
	forwardOK      int64
	forwardRetried int64
	forwardFailed  int64
	forwardDropped int64
	forwardBytes   int64
}

type profileStatusBucket struct {
	count int64
	bytes int64
}

type profileObservabilitySummary struct {
	mu sync.Mutex

	received   map[string]profileStatusBucket
	parseError map[string]int64
	profiles   map[profileSummaryKey]profileSummaryBucket
}

func newProfileObservabilitySummary() *profileObservabilitySummary {
	return &profileObservabilitySummary{
		received:   map[string]profileStatusBucket{},
		parseError: map[string]int64{},
		profiles:   map[profileSummaryKey]profileSummaryBucket{},
	}
}

func profileMetricsSetup() {
	profileReceivedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "received_total",
			Help:      "Profile intake requests received by DataKit.",
		},
		[]string{"status"},
	)

	profileReceivedBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "received_bytes_total",
			Help:      "Profile intake request body bytes received by DataKit.",
		},
		[]string{"status"},
	)

	profileParsedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "parsed_total",
			Help:      "Profile requests with metadata parsed by DataKit.",
		},
		[]string{"language", "format", "profiler"},
	)

	profileParsedBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "parsed_bytes_total",
			Help:      "Profile attachment bytes with metadata parsed by DataKit.",
		},
		[]string{"language", "format", "profiler"},
	)

	profileParseErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "parse_errors_total",
			Help:      "Profile parse failures by reason.",
		},
		[]string{"reason"},
	)

	profileForwardTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "forward_total",
			Help:      "Profile forwarding results to DataWay.",
		},
		[]string{"language", "format", "profiler", "status"},
	)

	profileForwardBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "profile",
			Name:      "forward_bytes_total",
			Help:      "Profile request body bytes forwarded to DataWay.",
		},
		[]string{"language", "format", "profiler", "status"},
	)

	profileForwardLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "profile",
			Name:       "forward_latency_seconds",
			Help:       "Profile forwarding latency to DataWay.",
			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"language", "format", "profiler", "status"},
	)
}

//nolint:gochecknoinits
func init() {
	profileMetricsSetup()

	cliMetrics.MustRegister(
		profileReceivedTotal,
		profileReceivedBytesTotal,
		profileParsedTotal,
		profileParsedBytesTotal,
		profileParseErrorsTotal,
		profileForwardTotal,
		profileForwardBytesTotal,
		profileForwardLatency,
	)
}

func (ipt *Input) startProfileObservabilityLogger() {
	if ipt.observabilityGroup != nil {
		return
	}

	ipt.observabilityGroup = goroutine.NewGroup(goroutine.Option{
		Name: "profile_observability",
		PanicCb: func(b []byte) bool {
			log.Errorf("goroutine profile-observability panic: %s", string(b))
			return false
		},
	})

	ipt.observabilityGroup.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-datakit.Exit.Wait():
				profileObsSummary.flushLog()
				return nil
			case <-ipt.semStop.Wait():
				profileObsSummary.flushLog()
				return nil
			case <-ctx.Done():
				profileObsSummary.flushLog()
				return nil
			case <-ticker.C:
				profileObsSummary.flushLog()
			}
		}
	})
}

func observeProfileReceived(status string, bytes int64) {
	status = normalizeProfileLabel(status)
	bytes = nonNegative(bytes)

	profileReceivedTotal.WithLabelValues(status).Inc()
	profileReceivedBytesTotal.WithLabelValues(status).Add(float64(bytes))
	profileObsSummary.recordReceived(status, bytes)
}

func observeProfileParsed(labels profileMetricLabels, bytes int64) {
	labels = labels.normalized()
	bytes = nonNegative(bytes)

	profileParsedTotal.WithLabelValues(labels.language, labels.format, labels.profiler).Inc()
	profileParsedBytesTotal.WithLabelValues(labels.language, labels.format, labels.profiler).Add(float64(bytes))
	profileObsSummary.recordParsed(labels, bytes)
}

func observeProfileParseError(reason string) {
	reason = normalizeProfileLabel(reason)

	profileParseErrorsTotal.WithLabelValues(reason).Inc()
	profileObsSummary.recordParseError(reason)
}

func observeProfileForward(labels profileMetricLabels, status string, bytes int64, latency time.Duration) {
	labels = labels.normalized()
	status = normalizeProfileLabel(status)
	bytes = nonNegative(bytes)

	profileForwardTotal.WithLabelValues(labels.language, labels.format, labels.profiler, status).Inc()
	profileForwardBytesTotal.WithLabelValues(labels.language, labels.format, labels.profiler, status).Add(float64(bytes))
	if latency >= 0 {
		profileForwardLatency.WithLabelValues(labels.language, labels.format, labels.profiler, status).
			Observe(latency.Seconds())
	}
	profileObsSummary.recordForward(labels, status, bytes)
}

func (s *profileObservabilitySummary) recordReceived(status string, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket := s.received[status]
	bucket.count++
	bucket.bytes += bytes
	s.received[status] = bucket
}

func (s *profileObservabilitySummary) recordParsed(labels profileMetricLabels, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := labels.summaryKey()
	bucket := s.profiles[key]
	bucket.parsedCount++
	bucket.parsedBytes += bytes
	s.profiles[key] = bucket
}

func (s *profileObservabilitySummary) recordParseError(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.parseError[reason]++
}

func (s *profileObservabilitySummary) recordForward(labels profileMetricLabels, status string, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := labels.summaryKey()
	bucket := s.profiles[key]
	switch status {
	case profileObsStatusOK:
		bucket.forwardOK++
		bucket.forwardBytes += bytes
	case profileObsStatusRetry:
		bucket.forwardRetried++
	case profileObsStatusDropped:
		bucket.forwardDropped++
		bucket.forwardBytes += bytes
	case profileObsStatusFailed:
		bucket.forwardFailed++
		bucket.forwardBytes += bytes
	}
	s.profiles[key] = bucket
}

func (s *profileObservabilitySummary) flushLog() {
	s.mu.Lock()
	if len(s.received) == 0 && len(s.parseError) == 0 && len(s.profiles) == 0 {
		s.mu.Unlock()
		return
	}

	received := s.received
	parseError := s.parseError
	profiles := s.profiles
	s.received = map[string]profileStatusBucket{}
	s.parseError = map[string]int64{}
	s.profiles = map[profileSummaryKey]profileSummaryBucket{}
	s.mu.Unlock()

	log.Infof("profile observability summary: received=%s parse_errors=%s profiles=%s",
		formatProfileReceivedSummary(received),
		formatProfileErrorSummary(parseError),
		formatProfileSummary(profiles),
	)
}

func formatProfileReceivedSummary(m map[string]profileStatusBucket) string {
	if len(m) == 0 {
		return "-"
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		bucket := m[key]
		parts = append(parts, fmt.Sprintf("%s:%d/%dB", key, bucket.count, bucket.bytes))
	}
	return strings.Join(parts, ",")
}

func formatProfileErrorSummary(m map[string]int64) string {
	if len(m) == 0 {
		return "-"
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", key, m[key]))
	}
	return strings.Join(parts, ",")
}

func formatProfileSummary(m map[profileSummaryKey]profileSummaryBucket) string {
	if len(m) == 0 {
		return "-"
	}

	keys := make([]profileSummaryKey, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].language != keys[j].language {
			return keys[i].language < keys[j].language
		}
		if keys[i].format != keys[j].format {
			return keys[i].format < keys[j].format
		}
		return keys[i].profiler < keys[j].profiler
	})

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		bucket := m[key]
		parts = append(parts, fmt.Sprintf(
			"%s/%s/%s:parsed=%d parsed_bytes=%d forward_ok=%d forward_retry=%d forward_failed=%d forward_dropped=%d forward_bytes=%d",
			key.language,
			key.format,
			key.profiler,
			bucket.parsedCount,
			bucket.parsedBytes,
			bucket.forwardOK,
			bucket.forwardRetried,
			bucket.forwardFailed,
			bucket.forwardDropped,
			bucket.forwardBytes,
		))
	}
	return strings.Join(parts, ";")
}

func resolveProfileMetricLabels(metadata map[string]string,
	files map[string][]*multipart.FileHeader,
) profileMetricLabels {
	return profileMetricLabels{
		language: resolveProfileLanguage(metadata),
		format:   resolveProfileFormat(metadata, files),
		profiler: resolveProfileProfiler(metadata),
	}.normalized()
}

func resolveProfileLanguage(metadata map[string]string) string {
	if metadata == nil {
		return profileObsUnknown
	}

	language := profileMetrics.ResolveLang(metadata).String()
	return normalizeProfileLabel(language)
}

func resolveProfileFormat(metadata map[string]string, files map[string][]*multipart.FileHeader) string {
	if metadata != nil {
		if format := metadata["format"]; format != "" {
			return normalizeProfileLabel(format)
		}
	}

	for field, headers := range files {
		if field == profileMetrics.EventFile || field == profileMetrics.EventJSONFile {
			continue
		}
		if format := formatFromProfileFileName(field); format != "" {
			return format
		}
		for _, header := range headers {
			if format := formatFromProfileFileName(header.Filename); format != "" {
				return format
			}
		}
	}

	return profileObsUnknown
}

func resolveProfileProfiler(metadata map[string]string) string {
	if metadata == nil {
		return profileObsUnknown
	}

	for _, key := range []string{"profiler", "library_type", "profiler_type"} {
		if profiler := metadata[key]; profiler != "" {
			return normalizeProfileLabel(profiler)
		}
	}

	if metadata["profiler_version"] != "" {
		return string(profileMetrics.DDtrace)
	}

	return profileObsUnknown
}

func formatFromProfileFileName(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pprof":
		return string(profileMetrics.PPROF)
	case ".jfr":
		return string(profileMetrics.JFR)
	default:
		return ""
	}
}

func normalizeProfileLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return profileObsUnknown
	}

	replacer := strings.NewReplacer(
		" ", "_",
		"\t", "_",
		"\n", "_",
		"\r", "_",
		",", "_",
		":", "_",
		"/", "_",
		"\\", "_",
	)
	return replacer.Replace(value)
}

func nonNegative(n int64) int64 {
	if n < 0 {
		return 0
	}
	return n
}

func (labels profileMetricLabels) normalized() profileMetricLabels {
	return profileMetricLabels{
		language: normalizeProfileLabel(labels.language),
		format:   normalizeProfileLabel(labels.format),
		profiler: normalizeProfileLabel(labels.profiler),
	}
}

func (labels profileMetricLabels) summaryKey() profileSummaryKey {
	labels = labels.normalized()
	return profileSummaryKey(labels)
}
