// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

var (
	jitRetirementTimeouts = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "datakit", Subsystem: "pipeline", Name: "jit_retirement_timeouts_total",
		Help: "Retired JIT generations exceeding the diagnostic drain deadline; in-use resources are retained.",
	})
	jitRetiredGenerations = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "datakit", Subsystem: "pipeline", Name: "jit_retired_generations",
		Help: "JIT generations awaiting final lease release or native cleanup.",
	}, func() float64 { jitRunners.mu.Lock(); defer jitRunners.mu.Unlock(); return float64(jitRunners.pending) })
	jitControlStateVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "datakit", Subsystem: "pipeline", Name: "jit_control_state",
		Help: "Current Pipeline JIT enabled or initial-degradation state.",
	}, []string{"state"})
	jitQuarantinesVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit", Subsystem: "pipeline", Name: "jit_quarantines_total",
		Help: "Pipeline JIT quarantine transitions by bounded script/runtime scope.",
	}, []string{"scope"})
	jitPhaseCostVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_phase_seconds",
		Help:      "Pipeline JIT batch duration split by phase.",
	}, []string{"category", "phase"})
	jitPhaseDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_phase_duration_seconds",
		Help:      "Pipeline JIT batch duration histogram split by phase for tail-latency tracking.",
		Buckets:   prometheus.ExponentialBuckets(0.00005, 2, 20),
	}, []string{"category", "phase"})
	jitBatchRecordsVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_batch_records",
		Help:      "Number of records submitted to each Pipeline JIT batch.",
	}, []string{"category"})
	jitBytesVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_bytes",
		Help:      "Pipeline JIT encoded input bytes for every submitted native batch and output bytes for native batches accepted by DataKit.",
	}, []string{"category", "direction", "protocol"})
	jitFallbackVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_fallback_records_total",
		Help:      "Compatibility counter for selected Pipeline JIT fallback reasons; use jit_route_records_total for complete routing outcomes.",
	}, []string{"category", "reason"})
	jitRouteRecordsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_route_records_total",
		Help:      "Pipeline records by bounded JIT routing stage and reason; attempted/submitted is a stage count and must not be summed with final outcomes.",
	}, []string{"category", "outcome", "reason"})
	jitHostCallbacksVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_host_callbacks_total",
		Help:      "Number of Pipeline JIT host ABI callback invocations by bounded operation; buffer retries count as additional invocations.",
	}, []string{"operation"})
	jitProgramChecksVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_program_checks_total",
		Help:      "Pipeline JIT route checks by bounded backend and execution tier; machine_code_helper still invokes Rust helpers, while machine_code_slots executes against native slots.",
	}, []string{"route", "reason", "backend", "execution_tier"})
	jitProgramCheckCostVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_program_check_seconds",
		Help:      "Pipeline JIT script route-check duration, including a first compile on cache miss.",
		Buckets:   prometheus.ExponentialBuckets(0.00025, 2, 15),
	}, []string{"route", "reason", "backend", "execution_tier"})
	jitRuntimeInitCostVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_runtime_init_seconds",
		Help:      "Pipeline JIT runtime library initialization duration.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"outcome"})
	jitAggregateDeliveryVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_aggregate_delivery_total",
		Help:      "Pipeline JIT aggregate upload attempts and points by bounded outcome; retries are separate attempts.",
	}, []string{"outcome", "unit"})
	jitAggregateDeliveryCostVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "datakit",
		Subsystem: "pipeline",
		Name:      "jit_aggregate_delivery_seconds",
		Help:      "Pipeline JIT aggregate conversion and upload duration by bounded outcome.",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
	}, []string{"outcome"})
	jitHostGeoIPCounter       = jitHostCallbacksVec.WithLabelValues("geoip")
	jitHostUserAgentCounter   = jitHostCallbacksVec.WithLabelValues("user_agent")
	jitHostDefaultTimeCounter = jitHostCallbacksVec.WithLabelValues("default_time")
	jitHostUnknownCounter     = jitHostCallbacksVec.WithLabelValues("unknown")
)

func jitMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		jitRetirementTimeouts,
		jitRetiredGenerations,
		jitControlStateVec,
		jitQuarantinesVec,
		jitPhaseCostVec,
		jitPhaseDurationVec,
		jitBatchRecordsVec,
		jitBytesVec,
		jitFallbackVec,
		jitRouteRecordsVec,
		jitHostCallbacksVec,
		jitProgramChecksVec,
		jitProgramCheckCostVec,
		jitRuntimeInitCostVec,
		jitAggregateDeliveryVec,
		jitAggregateDeliveryCostVec,
	}
}

func observeJITAggregateDelivery(points int, elapsed time.Duration, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	jitAggregateDeliveryVec.WithLabelValues(outcome, "attempts").Inc()
	jitAggregateDeliveryVec.WithLabelValues(outcome, "points").Add(float64(points))
	jitAggregateDeliveryCostVec.WithLabelValues(outcome).Observe(elapsed.Seconds())
}

func observeJITCheck(result pljit.CheckResult) {
	jitProgramChecksVec.WithLabelValues(jitCheckLabels(result)...).Inc()
}

func observeJITCheckDuration(result pljit.CheckResult, elapsed time.Duration) {
	jitProgramCheckCostVec.WithLabelValues(jitCheckLabels(result)...).Observe(elapsed.Seconds())
}

func jitCheckLabels(result pljit.CheckResult) []string {
	return []string{
		boundedJITCapabilityLabel(string(result.Route), string(pljit.RouteJITNative),
			string(pljit.RouteJITWithHost), string(pljit.RoutePipelineGo)),
		boundedJITCapabilityLabel(result.Reason, pljit.CheckReasonJITReady,
			pljit.CheckReasonRunnerUnavailable, pljit.CheckReasonCompileError,
			pljit.CheckReasonCapabilitiesError),
		boundedJITCapabilityLabel(result.Capabilities.Backend,
			pljit.ExecutionBackendInstructionEngine, pljit.ExecutionBackendMachineCode),
		boundedJITCapabilityLabel(result.Capabilities.ExecutionTier,
			pljit.ExecutionTierInstructionEngine,
			pljit.ExecutionTierMachineCodeHelper,
			pljit.ExecutionTierMachineCodeSlots),
	}
}

func observeJITRuntimeInit(outcome string, elapsed time.Duration) {
	jitRuntimeInitCostVec.WithLabelValues(boundedJITCapabilityLabel(outcome, "ready", "error")).Observe(elapsed.Seconds())
}

func boundedJITCapabilityLabel(value string, allowed ...string) string {
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return "unknown"
}

func observeJITHostCallback(operation string) {
	switch operation {
	case "geoip":
		jitHostGeoIPCounter.Inc()
	case "user_agent":
		jitHostUserAgentCounter.Inc()
	case "default_time":
		jitHostDefaultTimeCounter.Inc()
	default:
		jitHostUnknownCounter.Inc()
	}
}

func observeJITRoute(category point.Category, outcome, reason string, records int) {
	if records > 0 {
		jitRouteRecordsVec.WithLabelValues(category.String(), outcome, reason).Add(float64(records))
	}
}

func observeJITPhase(category point.Category, phase string, elapsed time.Duration) {
	jitPhaseCostVec.WithLabelValues(category.String(), phase).Observe(elapsed.Seconds())
	jitPhaseDurationVec.WithLabelValues(category.String(), phase).Observe(elapsed.Seconds())
}

func observeJITFallback(category point.Category, reason string, records int) {
	if records > 0 {
		jitFallbackVec.WithLabelValues(category.String(), reason).Add(float64(records))
	}
}

// Metric vectors are process-lifetime collectors. Resolve the bounded labels
// once per category so the execution path does not allocate label slices.
var jitGroupMetricsByCategory = new(sync.Map)

type jitPhaseObservers struct {
	cost     prometheus.Observer
	duration prometheus.Observer
}

func (p jitPhaseObservers) observe(elapsed time.Duration) {
	seconds := elapsed.Seconds()
	p.cost.Observe(seconds)
	p.duration.Observe(seconds)
}

type jitGroupMetrics struct {
	encode, native, apply                  jitPhaseObservers
	submitted, ok, dropped, committedError prometheus.Counter
	inputBytes, batchRecords               prometheus.Observer
	staticOutputBytes                      prometheus.Observer
	dynamicOutputBytes                     prometheus.Observer
}

func groupJITMetrics(category point.Category) *jitGroupMetrics {
	if cached, ok := jitGroupMetricsByCategory.Load(category); ok {
		return cached.(*jitGroupMetrics)
	}
	name := category.String()
	phase := func(value string) jitPhaseObservers {
		return jitPhaseObservers{jitPhaseCostVec.WithLabelValues(name, value), jitPhaseDurationVec.WithLabelValues(name, value)}
	}
	metrics := &jitGroupMetrics{
		encode: phase("encode"), native: phase("native"), apply: phase("apply"),
		submitted:          jitRouteRecordsVec.WithLabelValues(name, "attempted", "submitted"),
		ok:                 jitRouteRecordsVec.WithLabelValues(name, "native", "ok"),
		dropped:            jitRouteRecordsVec.WithLabelValues(name, "native", "dropped"),
		committedError:     jitRouteRecordsVec.WithLabelValues(name, "native", "committed_error"),
		inputBytes:         jitBytesVec.WithLabelValues(name, "input", "flat-v1"),
		staticOutputBytes:  jitBytesVec.WithLabelValues(name, "output", "static-v2"),
		dynamicOutputBytes: jitBytesVec.WithLabelValues(name, "output", "dynamic-v3"),
		batchRecords:       jitBatchRecordsVec.WithLabelValues(name),
	}
	actual, _ := jitGroupMetricsByCategory.LoadOrStore(category, metrics)
	return actual.(*jitGroupMetrics)
}
