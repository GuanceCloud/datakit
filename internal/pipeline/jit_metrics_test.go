// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func readPrometheusCounter(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		t.Fatalf("read counter: %v", err)
	}
	return metric.GetCounter().GetValue()
}

func TestObserveJITCheckUsesBoundedBackendAndTier(t *testing.T) {
	counter := jitProgramChecksVec.WithLabelValues(
		string(pljit.RouteJITNative), pljit.CheckReasonJITReady,
		pljit.ExecutionBackendMachineCode, pljit.ExecutionTierMachineCodeHelper,
	)
	before := readPrometheusCounter(t, counter)
	observeJITCheck(pljit.CheckResult{
		Route:  pljit.RouteJITNative,
		Reason: pljit.CheckReasonJITReady,
		Capabilities: pljit.ProgramCapabilities{
			Backend:       pljit.ExecutionBackendMachineCode,
			ExecutionTier: pljit.ExecutionTierMachineCodeHelper,
		},
	})
	if delta := readPrometheusCounter(t, counter) - before; delta != 1 {
		t.Fatalf("program check counter delta = %v, want 1", delta)
	}

	unknown := jitProgramChecksVec.WithLabelValues("unknown", "unknown", "unknown", "unknown")
	before = readPrometheusCounter(t, unknown)
	observeJITCheck(pljit.CheckResult{
		Route:  "remote_jit",
		Reason: "source-controlled-reason",
		Capabilities: pljit.ProgramCapabilities{
			Backend:       "source-controlled-backend",
			ExecutionTier: "source-controlled-tier",
		},
	})
	if delta := readPrometheusCounter(t, unknown) - before; delta != 1 {
		t.Fatalf("unknown program check counter delta = %v, want 1", delta)
	}
}

func readPrometheusHistogram(t *testing.T, observer prometheus.Observer) (uint64, float64) {
	t.Helper()
	metric, ok := observer.(prometheus.Metric)
	if !ok {
		t.Fatalf("observer %T is not a Prometheus metric", observer)
	}
	encoded := &dto.Metric{}
	if err := metric.Write(encoded); err != nil {
		t.Fatalf("read histogram: %v", err)
	}
	return encoded.GetHistogram().GetSampleCount(), encoded.GetHistogram().GetSampleSum()
}

func TestObserveJITStartupDurationsUseBoundedLabels(t *testing.T) {
	result := pljit.CheckResult{
		Route:  pljit.RouteJITNative,
		Reason: pljit.CheckReasonJITReady,
		Capabilities: pljit.ProgramCapabilities{
			Backend:       pljit.ExecutionBackendMachineCode,
			ExecutionTier: pljit.ExecutionTierMachineCodeSlots,
		},
	}
	check := jitProgramCheckCostVec.WithLabelValues(
		string(pljit.RouteJITNative), pljit.CheckReasonJITReady,
		pljit.ExecutionBackendMachineCode, pljit.ExecutionTierMachineCodeSlots,
	)
	beforeCount, beforeSum := readPrometheusHistogram(t, check)
	observeJITCheckDuration(result, 7*time.Millisecond)
	afterCount, afterSum := readPrometheusHistogram(t, check)
	if afterCount-beforeCount != 1 || math.Abs(afterSum-beforeSum-0.007) > 1e-12 {
		t.Fatalf("program check histogram delta = (%d, %v), want (1, 0.007)", afterCount-beforeCount, afterSum-beforeSum)
	}

	unknown := jitRuntimeInitCostVec.WithLabelValues("unknown")
	beforeCount, beforeSum = readPrometheusHistogram(t, unknown)
	observeJITRuntimeInit("source-controlled-outcome", 3*time.Millisecond)
	afterCount, afterSum = readPrometheusHistogram(t, unknown)
	if afterCount-beforeCount != 1 || math.Abs(afterSum-beforeSum-0.003) > 1e-12 {
		t.Fatalf("runtime init histogram delta = (%d, %v), want (1, 0.003)", afterCount-beforeCount, afterSum-beforeSum)
	}
}

func TestObserveJITPhaseRecordsTailLatencyHistogram(t *testing.T) {
	phase := jitPhaseDurationVec.WithLabelValues(point.Logging.String(), "native")
	beforeCount, beforeSum := readPrometheusHistogram(t, phase)
	observeJITPhase(point.Logging, "native", 25*time.Millisecond)
	afterCount, afterSum := readPrometheusHistogram(t, phase)
	if afterCount-beforeCount != 1 || math.Abs(afterSum-beforeSum-0.025) > 1e-12 {
		t.Fatalf("phase histogram delta = (%d, %v), want (1, 0.025)", afterCount-beforeCount, afterSum-beforeSum)
	}
}

func TestObserveJITAggregateDeliveryRecordsRetriesAndPoints(t *testing.T) {
	errorAttempts := jitAggregateDeliveryVec.WithLabelValues("error", "attempts")
	errorPoints := jitAggregateDeliveryVec.WithLabelValues("error", "points")
	beforeAttempts := readPrometheusCounter(t, errorAttempts)
	beforePoints := readPrometheusCounter(t, errorPoints)
	beforeCount, _ := readPrometheusHistogram(t, jitAggregateDeliveryCostVec.WithLabelValues("error"))

	observeJITAggregateDelivery(7, 2*time.Millisecond, fmt.Errorf("upload unavailable"))

	if delta := readPrometheusCounter(t, errorAttempts) - beforeAttempts; delta != 1 {
		t.Fatalf("aggregate error attempt delta=%v, want 1", delta)
	}
	if delta := readPrometheusCounter(t, errorPoints) - beforePoints; delta != 7 {
		t.Fatalf("aggregate error point delta=%v, want 7", delta)
	}
	afterCount, _ := readPrometheusHistogram(t, jitAggregateDeliveryCostVec.WithLabelValues("error"))
	if afterCount-beforeCount != 1 {
		t.Fatalf("aggregate duration count delta=%d, want 1", afterCount-beforeCount)
	}
}

func readPrometheusSummary(t *testing.T, observer prometheus.Observer) (uint64, float64) {
	t.Helper()
	metric, ok := observer.(prometheus.Metric)
	if !ok {
		t.Fatalf("observer %T is not a Prometheus metric", observer)
	}
	encoded := &dto.Metric{}
	if err := metric.Write(encoded); err != nil {
		t.Fatalf("read summary: %v", err)
	}
	return encoded.GetSummary().GetSampleCount(), encoded.GetSummary().GetSampleSum()
}

func TestObserveJITRoute(t *testing.T) {
	counter := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "attempted", "submitted")
	before := readPrometheusCounter(t, counter)
	observeJITRoute(point.Logging, "attempted", "submitted", 3)
	observeJITRoute(point.Logging, "attempted", "submitted", 0)
	if delta := readPrometheusCounter(t, counter) - before; delta != 3 {
		t.Fatalf("route counter delta = %v, want 3", delta)
	}
}

func TestObserveJITHostCallbackUsesBoundedLabels(t *testing.T) {
	tests := []struct {
		operation string
		counter   prometheus.Counter
	}{
		{"geoip", jitHostGeoIPCounter},
		{"user_agent", jitHostUserAgentCounter},
		{"default_time", jitHostDefaultTimeCounter},
		{"future-operation", jitHostUnknownCounter},
	}
	for _, test := range tests {
		before := readPrometheusCounter(t, test.counter)
		observeJITHostCallback(test.operation)
		if delta := readPrometheusCounter(t, test.counter) - before; delta != 1 {
			t.Fatalf("host callback %q counter delta = %v, want 1", test.operation, delta)
		}
	}
}

func TestJITMetricsRegisterHostCallbackCollector(t *testing.T) {
	registry := prometheus.NewPedanticRegistry()
	for _, collector := range jitMetrics() {
		if err := registry.Register(collector); err != nil {
			t.Fatalf("register JIT metric: %v", err)
		}
	}
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather JIT metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "datakit_pipeline_jit_host_callbacks_total" {
			continue
		}
		operations := make(map[string]bool)
		for _, metric := range family.Metric {
			for _, label := range metric.Label {
				if label.GetName() == "operation" {
					operations[label.GetValue()] = true
				}
			}
		}
		for _, operation := range []string{"geoip", "user_agent", "default_time", "unknown"} {
			if !operations[operation] {
				t.Fatalf("host callback metric missing bounded operation %q: %v", operation, operations)
			}
		}
		if len(operations) != 4 {
			t.Fatalf("host callback metric has unbounded operations: %v", operations)
		}
		return
	}
	t.Fatal("datakit_pipeline_jit_host_callbacks_total was not registered")
}
