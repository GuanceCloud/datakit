// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"context"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

const kubeletStatsSummarySource = "kubelet_stats_summary"

// resourceTriggerState keeps a sliding window of violating samples per metric.
// Samples are recorded on each monitor evaluation and aged out of the window,
// so transient dips only weaken the window instead of resetting it, and a
// skipped evaluation (e.g. a kubelet stats error) does not discard history.
type resourceTriggerState struct {
	violations map[string][]time.Time
}

type resourceUsage struct {
	cpuMillicores int64
	memoryBytes   int64
	cpuLimit      int64
	memoryLimit   int64
}

type resourceTrigger struct {
	metric    string
	value     float64
	threshold float64
	unit      string
	emergency bool
}

func (t *resourceTrigger) tags() map[string]string {
	level := "normal"
	if t.emergency {
		level = "emergency"
	}
	return map[string]string{
		"trigger_metric":    t.metric,
		"trigger_value":     strconv.FormatFloat(t.value, 'f', -1, 64),
		"trigger_threshold": strconv.FormatFloat(t.threshold, 'f', -1, 64),
		"trigger_unit":      t.unit,
		"trigger_source":    kubeletStatsSummarySource,
		"trigger_level":     level,
	}
}

type triggerCandidate struct {
	metric    string
	value     float64
	threshold float64
	unit      string
	emergency bool
	available bool
}

func (t *kubernetesProfileTarget) evaluateResourceTrigger(usage resourceUsage, now time.Time) *resourceTrigger {
	cfg := t.rule.config
	cpuPercent, cpuPercentAvailable := usagePercent(usage.cpuMillicores, usage.cpuLimit)
	memoryPercent, memoryPercentAvailable := usagePercent(usage.memoryBytes, usage.memoryLimit)
	candidates := []triggerCandidate{
		{"cpu_usage_base_limit", cpuPercent, cfg.CPUUsageBaseLimit, "percent", false, cpuPercentAvailable},
		{"cpu_usage_millicores", float64(usage.cpuMillicores), float64(cfg.CPUUsageMillicores), "millicores", false, true},
		{"mem_usage_base_limit", memoryPercent, cfg.MemUsageBaseLimit, "percent", false, memoryPercentAvailable},
		{"mem_usage_bytes", float64(usage.memoryBytes), float64(cfg.MemUsageBytes), "bytes", false, true},
		{"cpu_emergency_base_limit", cpuPercent, cfg.CPUEmergencyBaseLimit, "percent", true, cpuPercentAvailable},
		{"cpu_emergency_millicores", float64(usage.cpuMillicores), float64(cfg.CPUEmergencyMillicores), "millicores", true, true},
		{"mem_emergency_base_limit", memoryPercent, cfg.MemEmergencyBaseLimit, "percent", true, memoryPercentAvailable},
		{"mem_emergency_bytes", float64(usage.memoryBytes), float64(cfg.MemEmergencyBytes), "bytes", true, true},
	}

	if t.triggerState.violations == nil {
		t.triggerState.violations = make(map[string][]time.Time)
	}

	// Emergency conditions bypass the sliding window.
	for _, candidate := range candidates {
		if candidate.emergency && candidate.exceeded() {
			return candidate.asTrigger()
		}
	}

	// Normal conditions use a sliding window over recent monitor samples: a
	// trigger fires once enough violating samples fall inside trigger_window.
	// Non-violating samples are not recorded, so a transient dip does not reset
	// the window, and samples are pruned once they age out of the window.
	windowStart := now.Add(-cfg.TriggerWindow)
	required := requiredTriggerSamples(cfg)
	var triggered *resourceTrigger
	for _, candidate := range candidates {
		if candidate.emergency || candidate.threshold <= 0 {
			continue
		}
		samples := t.triggerState.violations[candidate.metric]
		keep := 0
		for _, sample := range samples {
			if !sample.Before(windowStart) {
				samples[keep] = sample
				keep++
			}
		}
		samples = samples[:keep]
		if candidate.exceeded() {
			samples = append(samples, now)
		}
		t.triggerState.violations[candidate.metric] = samples
		if triggered == nil && len(samples) >= required {
			triggered = candidate.asTrigger()
		}
	}
	return triggered
}

// requiredTriggerSamples returns how many violating samples must fall inside
// trigger_window before a normal threshold fires.
func requiredTriggerSamples(cfg *KubernetesProfiler) int {
	count := int(cfg.TriggerWindow / cfg.MonitorInterval)
	if count < 1 {
		return 1
	}
	return count
}

func (c triggerCandidate) exceeded() bool {
	return c.available && c.threshold > 0 && c.value >= c.threshold
}

func (c triggerCandidate) asTrigger() *resourceTrigger {
	return &resourceTrigger{
		metric:    c.metric,
		value:     c.value,
		threshold: c.threshold,
		unit:      c.unit,
		emergency: c.emergency,
	}
}

func usagePercent(usage, limit int64) (float64, bool) {
	if limit <= 0 {
		return 0, false
	}
	return float64(usage) / float64(limit) * 100, true
}

type containerStatsKey struct {
	namespace string
	podName   string
	podUID    string
	container string
}

type containerResourceStats struct {
	cpuMillicores int64
	memoryBytes   int64
}

func indexContainerStats(summary *statsv1alpha1.Summary) map[containerStatsKey]containerResourceStats {
	result := make(map[containerStatsKey]containerResourceStats)
	if summary == nil {
		return result
	}
	for _, pod := range summary.Pods {
		for _, container := range pod.Containers {
			key := containerStatsKey{
				namespace: pod.PodRef.Namespace,
				podName:   pod.PodRef.Name,
				podUID:    pod.PodRef.UID,
				container: container.Name,
			}
			stats := containerResourceStats{}
			if container.CPU != nil && container.CPU.UsageNanoCores != nil {
				stats.cpuMillicores = int64(*container.CPU.UsageNanoCores) / 1e6
			}
			if container.Memory != nil && container.Memory.WorkingSetBytes != nil {
				stats.memoryBytes = int64(*container.Memory.WorkingSetBytes)
			}
			result[key] = stats
		}
	}
	return result
}

func findContainerStats(
	index map[containerStatsKey]containerResourceStats,
	target *kubernetesProfileTarget,
) (containerResourceStats, bool) {
	parts := splitPodKey(target.podKey)
	if len(parts) != 2 {
		return containerResourceStats{}, false
	}
	key := containerStatsKey{
		namespace: parts[0],
		podName:   parts[1],
		podUID:    target.podUID,
		container: target.container,
	}
	if stats, ok := index[key]; ok {
		return stats, true
	}
	// Older kubelets may omit Pod UID in the summary.
	key.podUID = ""
	stats, ok := index[key]
	return stats, ok
}

func splitPodKey(key string) []string {
	for i := range key {
		if key[i] == '/' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return nil
}

func (m *kubernetesProfileManager) runMonitorLoop(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			m.monitorWithContext(ctx, time.Now())
			timer.Reset(m.nextMonitorDelay(time.Now()))
		}
	}
}

func (m *kubernetesProfileManager) nextMonitorDelay(now time.Time) time.Duration {
	delay := m.minimumMonitorInterval()
	for _, target := range m.targetSnapshot() {
		target.mu.Lock()
		if target.ready && !target.removed && hasResourceTrigger(target.rule.config) {
			if remaining := target.nextMonitor.Sub(now); remaining < delay {
				delay = remaining
			}
		}
		target.mu.Unlock()
	}
	if delay < 0 {
		return 0
	}
	return delay
}

func (m *kubernetesProfileManager) minimumMonitorInterval() time.Duration {
	minimum := defaultKubernetesProfileMonitorInterval
	found := false
	for _, rule := range m.rules {
		if !hasResourceTrigger(rule.config) {
			continue
		}
		if !found || rule.config.MonitorInterval < minimum {
			minimum = rule.config.MonitorInterval
			found = true
		}
	}
	return minimum
}

func (m *kubernetesProfileManager) monitor(now time.Time) {
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	m.monitorWithContext(ctx, now)
}

func (m *kubernetesProfileManager) monitorWithContext(ctx context.Context, now time.Time) {
	targets := m.dueMonitorTargets(now)
	if len(targets) == 0 {
		return
	}

	timeout := m.minimumMonitorInterval()
	if timeout > 10*time.Second {
		timeout = 10 * time.Second
	}
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	summary, err := m.stats.GetStatsSummaryWithContext(requestCtx)
	if err != nil {
		observeKubernetesProfileSkipped("stats_error")
		log.Warnf("query kubelet stats/summary for Kubernetes profile triggers failed: %s", err)
		return
	}
	index := indexContainerStats(summary)
	for _, target := range targets {
		target.mu.Lock()
		if target.removed || !target.ready {
			target.mu.Unlock()
			continue
		}
		stats, ok := findContainerStats(index, target)
		if !ok {
			target.mu.Unlock()
			observeKubernetesProfileSkipped("stats_not_found")
			continue
		}
		cpuLimit, memoryLimit := podutil.ContainerLimitInPod(target.container, target.pod)
		trigger := target.evaluateResourceTrigger(resourceUsage{
			cpuMillicores: stats.cpuMillicores,
			memoryBytes:   stats.memoryBytes,
			cpuLimit:      cpuLimit,
			memoryLimit:   memoryLimit,
		}, now)
		target.mu.Unlock()
		if trigger == nil {
			continue
		}

		duration := target.rule.config.ProfileDuration
		if trigger.emergency {
			duration = target.rule.config.EmergencyDuration
		}
		m.dispatch(target, "resource_threshold", target.rule.config.TriggerTypes, duration, trigger, now)
	}
}

func (m *kubernetesProfileManager) dueMonitorTargets(now time.Time) []*kubernetesProfileTarget {
	var due []*kubernetesProfileTarget
	for _, target := range m.targetSnapshot() {
		target.mu.Lock()
		if target.removed || !target.ready || !hasResourceTrigger(target.rule.config) || now.Before(target.nextMonitor) {
			target.mu.Unlock()
			continue
		}
		interval := target.rule.config.MonitorInterval
		remaining := interval - now.Sub(target.nextMonitor)%interval
		target.nextMonitor = now.Add(remaining)
		target.mu.Unlock()
		due = append(due, target)
	}
	return due
}
