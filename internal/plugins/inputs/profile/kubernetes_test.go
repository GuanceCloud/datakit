// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type fakeProfileStatsProvider struct {
	summary *statsv1alpha1.Summary
	err     error
}

func (p *fakeProfileStatsProvider) GetStatsSummaryWithContext(context.Context) (*statsv1alpha1.Summary, error) {
	return p.summary, p.err
}

type profileCollectCall struct {
	types    []string
	duration time.Duration
	tags     map[string]string
}

type fakeProfileCollector struct {
	mu      sync.Mutex
	calls   []profileCollectCall
	started chan struct{}
	release chan struct{}
	err     error
}

func (c *fakeProfileCollector) collect(
	ctx context.Context,
	types []string,
	duration time.Duration,
	tags map[string]string,
) error {
	if c.started != nil {
		c.started <- struct{}{}
	}
	if c.release != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.release:
		}
	}
	c.mu.Lock()
	c.calls = append(c.calls, profileCollectCall{
		types:    append([]string(nil), types...),
		duration: duration,
		tags:     copyTags(tags),
	})
	c.mu.Unlock()
	return c.err
}

func (c *fakeProfileCollector) snapshot() []profileCollectCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]profileCollectCall(nil), c.calls...)
}

func TestDecodeKubernetesProfileConfig(t *testing.T) {
	var decoded struct {
		Profile []*Input `toml:"profile"`
	}
	_, err := bstoml.Decode(`
[[profile]]
  [[profile.kubernetes]]
    node_local = true
    namespaces = ["guancedb"]
    selector = "app=select"
    port = "pprof"
    interval = "10m"
    profile_duration = "30s"
    monitor_interval = "10s"
    trigger_window = "1m"
    cooldown = "10m"
    cpu_usage_base_limit = 80
`, &decoded)
	require.NoError(t, err)
	require.Len(t, decoded.Profile, 1)
	require.Len(t, decoded.Profile[0].Kubernetes, 1)
	cfg := decoded.Profile[0].Kubernetes[0]
	require.NotNil(t, cfg.NodeLocal)
	assert.True(t, *cfg.NodeLocal)
	assert.Equal(t, 10*time.Minute, cfg.Interval)
	assert.Equal(t, 30*time.Second, cfg.ProfileDuration)

	rules, err := normalizeKubernetesProfileRules(decoded.Profile[0].Kubernetes)
	require.NoError(t, err)
	assert.Equal(t, []string{"heap", "goroutine"}, rules[0].config.ScheduledTypes)
	assert.Equal(t, []string{"cpu", "heap", "goroutine"}, rules[0].config.TriggerTypes)
	assert.Equal(t, 2, rules[0].config.MaxConcurrency)
}

func TestNormalizeKubernetesProfileRulesRejectsInvalidConfig(t *testing.T) {
	falseValue := false
	tests := []struct {
		name string
		cfg  *KubernetesProfiler
	}{
		{"cluster wide", &KubernetesProfiler{NodeLocal: &falseValue, Port: "6060"}},
		{"missing port", &KubernetesProfiler{}},
		{"invalid selector", &KubernetesProfiler{Port: "6060", Selector: "a in ("}},
		{"invalid profile type", &KubernetesProfiler{Port: "6060", ScheduledTypes: []string{"trace"}}},
		{"timeout shorter than CPU", &KubernetesProfiler{Port: "6060", ProfileDuration: 30 * time.Second, RequestTimeout: 10 * time.Second}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizeKubernetesProfileRules([]*KubernetesProfiler{test.cfg})
			require.Error(t, err)
		})
	}
}

func TestKubernetesProfileDesiredTargets(t *testing.T) {
	t.Setenv("ENV_CLUSTER_NAME_K8S", "test-cluster")
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{
		{
			Namespaces: []string{"guancedb"},
			Selector:   "app=select",
			Container:  "select",
			Port:       "pprof",
			PodLabelAsTags: map[string]string{
				"team": "owner_team",
			},
		},
	})
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	desired := manager.desiredTargets(pod)
	require.Len(t, desired, 1)
	for _, target := range desired {
		assert.Equal(t, "http://10.0.0.8:6060/debug/pprof", target.endpoint)
		assert.Equal(t, "select", target.container)
		assert.Equal(t, "guancedb", target.tags["namespace"])
		assert.Equal(t, "test-cluster", target.tags["cluster_name_k8s"])
		assert.Equal(t, "select", target.tags["service"])
		assert.Equal(t, "v1", target.tags["version"])
		assert.Equal(t, "database", target.tags["owner_team"])
		assert.Equal(t, "deployment", target.tags["workload_kind"])
		assert.Equal(t, "select", target.tags["workload_name"])
	}

	pod.Spec.NodeName = "node-b"
	assert.Empty(t, manager.desiredTargets(pod))
}

func TestKubernetesProfileReconcileLifecycle(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
	collector := &fakeProfileCollector{}
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return collector, nil
	}

	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	manager.reconcilePod(pod)
	require.Len(t, manager.targets, 1)
	var original *kubernetesProfileTarget
	for _, target := range manager.targets {
		original = target
	}

	pod.Labels["app.kubernetes.io/version"] = "v2"
	manager.reconcilePod(pod)
	require.Len(t, manager.targets, 1)
	for _, target := range manager.targets {
		assert.Same(t, original, target)
		assert.Equal(t, "v2", target.tags["version"])
	}

	pod.Status.PodIP = "10.0.0.9"
	manager.reconcilePod(pod)
	require.Len(t, manager.targets, 1)
	for _, target := range manager.targets {
		assert.NotSame(t, original, target)
		assert.Equal(t, "http://10.0.0.9:6060/debug/pprof", target.endpoint)
	}

	manager.deleteObject(pod)
	assert.Empty(t, manager.targets)
}

func TestKubernetesProfileReconcileReplacesChangedRule(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{
		{Port: "pprof", Selector: "app.kubernetes.io/version=v1", Service: "v1-service"},
		{Port: "pprof", Selector: "app.kubernetes.io/version=v2", Service: "v2-service"},
	})
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return &fakeProfileCollector{}, nil
	}
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	manager.reconcilePod(pod)
	require.Len(t, manager.targets, 1)
	var original *kubernetesProfileTarget
	for _, target := range manager.targets {
		original = target
		assert.Equal(t, 0, target.rule.index)
	}

	pod.Labels["app.kubernetes.io/version"] = "v2"
	manager.reconcilePod(pod)
	require.Len(t, manager.targets, 1)
	for _, target := range manager.targets {
		assert.NotSame(t, original, target)
		assert.Equal(t, 1, target.rule.index)
		assert.Equal(t, "v2-service", target.tags["service"])
	}
}

func TestResourceTriggerWindowAndEmergency(t *testing.T) {
	cfg := &KubernetesProfiler{
		Port:                  "6060",
		MonitorInterval:       time.Second,
		TriggerWindow:         2 * time.Second,
		CPUUsageBaseLimit:     80,
		MemEmergencyBaseLimit: 95,
	}
	rules, err := normalizeKubernetesProfileRules([]*KubernetesProfiler{cfg})
	require.NoError(t, err)
	target := &kubernetesProfileTarget{rule: rules[0]}
	now := time.Now()
	usage := resourceUsage{cpuMillicores: 900, cpuLimit: 1000, memoryBytes: 50, memoryLimit: 100}
	assert.Nil(t, target.evaluateResourceTrigger(usage, now))
	// trigger_window=2s with monitor_interval=1s requires two violating samples.
	trigger := target.evaluateResourceTrigger(usage, now.Add(time.Second))
	require.NotNil(t, trigger)
	assert.Equal(t, "cpu_usage_base_limit", trigger.metric)
	assert.False(t, trigger.emergency)

	usage.memoryBytes = 96
	trigger = target.evaluateResourceTrigger(usage, now.Add(2*time.Second))
	require.NotNil(t, trigger)
	assert.Equal(t, "mem_emergency_base_limit", trigger.metric)
	assert.True(t, trigger.emergency)
	assert.Equal(t, kubeletStatsSummarySource, trigger.tags()["trigger_source"])
}

func TestResourceTriggerSlidingWindowToleratesDips(t *testing.T) {
	cfg := &KubernetesProfiler{
		Port:              "6060",
		MonitorInterval:   time.Second,
		TriggerWindow:     3 * time.Second,
		CPUUsageBaseLimit: 80,
	}
	rules, err := normalizeKubernetesProfileRules([]*KubernetesProfiler{cfg})
	require.NoError(t, err)
	target := &kubernetesProfileTarget{rule: rules[0]}
	now := time.Now()
	hot := resourceUsage{cpuMillicores: 900, cpuLimit: 1000}
	cool := resourceUsage{cpuMillicores: 100, cpuLimit: 1000}

	assert.Nil(t, target.evaluateResourceTrigger(hot, now))
	assert.Nil(t, target.evaluateResourceTrigger(hot, now.Add(time.Second)))
	// A transient dip must not reset the sliding window.
	assert.Nil(t, target.evaluateResourceTrigger(cool, now.Add(2*time.Second)))
	trigger := target.evaluateResourceTrigger(hot, now.Add(3*time.Second))
	require.NotNil(t, trigger)
	assert.Equal(t, "cpu_usage_base_limit", trigger.metric)
}

func TestResourceTriggerAbsoluteThresholdsWithoutLimit(t *testing.T) {
	cfg := &KubernetesProfiler{
		Port:               "6060",
		MonitorInterval:    time.Second,
		TriggerWindow:      time.Second,
		CPUUsageMillicores: 500,
		MemUsageBytes:      100,
	}
	rules, err := normalizeKubernetesProfileRules([]*KubernetesProfiler{cfg})
	require.NoError(t, err)
	target := &kubernetesProfileTarget{rule: rules[0]}
	// No limits: percentage candidates must be unavailable, absolute ones must fire.
	trigger := target.evaluateResourceTrigger(resourceUsage{cpuMillicores: 600, memoryBytes: 150}, time.Now())
	require.NotNil(t, trigger)
	assert.Equal(t, "cpu_usage_millicores", trigger.metric)
	assert.Equal(t, "millicores", trigger.tags()["trigger_unit"])
}

func TestKubernetesProfileMonitorDispatchesResourceTrigger(t *testing.T) {
	usageNanoCores := uint64(900 * 1e6)
	workingSetBytes := uint64(90)
	stats := &fakeProfileStatsProvider{summary: &statsv1alpha1.Summary{
		Pods: []statsv1alpha1.PodStats{
			{
				PodRef: statsv1alpha1.PodReference{Name: "select-7fdc", Namespace: "guancedb", UID: "uid-a"},
				Containers: []statsv1alpha1.ContainerStats{
					{
						Name:   "select",
						CPU:    &statsv1alpha1.CPUStats{UsageNanoCores: &usageNanoCores},
						Memory: &statsv1alpha1.MemoryStats{WorkingSetBytes: &workingSetBytes},
					},
				},
			},
		},
	}}
	manager := newTestKubernetesProfileManagerWithStats(t, []*KubernetesProfiler{
		{
			Port:              "pprof",
			MonitorInterval:   time.Second,
			TriggerWindow:     time.Second,
			CPUUsageBaseLimit: 80,
			TriggerTypes:      []string{"goroutine"},
			ProfileDuration:   4 * time.Second,
			Cooldown:          time.Minute,
		},
	}, stats)
	collector := &fakeProfileCollector{}
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return collector, nil
	}
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))

	now := time.Now().Add(time.Second)
	manager.monitor(now)
	manager.monitor(now.Add(time.Second))
	manager.workers.Wait()
	calls := collector.snapshot()
	require.Len(t, calls, 1)
	assert.Equal(t, []string{"goroutine"}, calls[0].types)
	assert.Equal(t, 4*time.Second, calls[0].duration)
	assert.Equal(t, "resource_threshold", calls[0].tags["trigger_type"])
	assert.Equal(t, "cpu_usage_base_limit", calls[0].tags["trigger_metric"])
	assert.Equal(t, "90", calls[0].tags["trigger_value"])

	manager.monitor(now.Add(2 * time.Second))
	manager.workers.Wait()
	assert.Len(t, collector.snapshot(), 1, "cooldown must suppress repeated triggers")
}

func TestKubernetesProfileRuleConcurrency(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof", MaxConcurrency: 1}})
	collector := &fakeProfileCollector{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return collector, nil
	}
	podA := testProfilePod("uid-a", "10.0.0.8", "v1")
	podB := testProfilePod("uid-b", "10.0.0.9", "v1")
	podB.Name = "select-b"
	manager.reconcilePod(podA)
	manager.reconcilePod(podB)
	targets := manager.targetSnapshot()
	require.Len(t, targets, 2)

	now := time.Now()
	require.True(t, manager.dispatch(targets[0], "scheduled", []string{"goroutine"}, time.Second, nil, now))
	select {
	case <-collector.started:
	case <-time.After(time.Second):
		t.Fatal("first collection did not start")
	}
	assert.False(t, manager.dispatch(targets[1], "scheduled", []string{"goroutine"}, time.Second, nil, now))
	close(collector.release)
	manager.workers.Wait()
	assert.Len(t, collector.snapshot(), 1)
}

func TestKubernetesProfileMonitorStatsErrorKeepsWindow(t *testing.T) {
	usageNanoCores := uint64(900 * 1e6)
	workingSetBytes := uint64(90)
	stats := &fakeProfileStatsProvider{summary: &statsv1alpha1.Summary{
		Pods: []statsv1alpha1.PodStats{
			{
				PodRef: statsv1alpha1.PodReference{Name: "select-7fdc", Namespace: "guancedb", UID: "uid-a"},
				Containers: []statsv1alpha1.ContainerStats{
					{
						Name:   "select",
						CPU:    &statsv1alpha1.CPUStats{UsageNanoCores: &usageNanoCores},
						Memory: &statsv1alpha1.MemoryStats{WorkingSetBytes: &workingSetBytes},
					},
				},
			},
		},
	}}
	manager := newTestKubernetesProfileManagerWithStats(t, []*KubernetesProfiler{
		{
			Port:              "pprof",
			MonitorInterval:   time.Second,
			TriggerWindow:     3 * time.Second,
			CPUUsageBaseLimit: 80,
			TriggerTypes:      []string{"goroutine"},
			ProfileDuration:   time.Second,
			Cooldown:          time.Minute,
		},
	}, stats)
	collector := &fakeProfileCollector{}
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return collector, nil
	}
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))

	now := time.Now().Add(time.Second)
	manager.monitor(now)
	manager.monitor(now.Add(time.Second))
	// A kubelet stats error must not reset the sliding window.
	stats.err = errors.New("kubelet unavailable")
	manager.monitor(now.Add(2 * time.Second))
	stats.err = nil
	manager.monitor(now.Add(3 * time.Second))
	manager.workers.Wait()
	require.Len(t, collector.snapshot(), 1)
	assert.Equal(t, "resource_threshold", collector.snapshot()[0].tags["trigger_type"])
}

func TestKubernetesProfileDispatchFailureBackoff(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
	collector := &fakeProfileCollector{err: errors.New("pprof unreachable")}
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return collector, nil
	}
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
	targets := manager.targetSnapshot()
	require.Len(t, targets, 1)
	target := targets[0]

	now := time.Now()
	require.True(t, manager.dispatch(target, "scheduled", []string{"goroutine"}, time.Second, nil, now))
	manager.workers.Wait()

	target.mu.Lock()
	assert.False(t, target.busy, "busy flag must be released after failure")
	assert.Equal(t, 1, target.failures)
	assert.True(t, target.backoffUntil.After(now), "failure must activate backoff")
	target.mu.Unlock()

	assert.False(t, manager.dispatch(target, "scheduled", []string{"goroutine"}, time.Second, nil, now),
		"dispatch within backoff window must be rejected")
}

type panicProfileCollector struct{}

func (panicProfileCollector) collect(context.Context, []string, time.Duration, map[string]string) error {
	panic("collector boom")
}

func TestKubernetesProfileDispatchRecoversFromCollectorPanic(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return panicProfileCollector{}, nil
	}
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
	targets := manager.targetSnapshot()
	require.Len(t, targets, 1)
	target := targets[0]

	require.True(t, manager.dispatch(target, "scheduled", []string{"goroutine"}, time.Second, nil, time.Now()))
	manager.workers.Wait()

	target.mu.Lock()
	assert.False(t, target.busy, "busy flag must be released after a recovered panic")
	assert.Equal(t, 1, target.failures)
	assert.False(t, target.backoffUntil.IsZero())
	target.mu.Unlock()

	// The rule semaphore must have been released by the panic recovery.
	require.True(t, manager.dispatch(target, "scheduled", []string{"goroutine"}, time.Second, nil, time.Now().Add(time.Minute)))
	manager.workers.Wait()
	target.mu.Lock()
	assert.Equal(t, 2, target.failures)
	target.mu.Unlock()
}

func TestIndexContainerStats(t *testing.T) {
	cpu := uint64(250 * 1e6)
	memory := uint64(1024)
	index := indexContainerStats(&statsv1alpha1.Summary{Pods: []statsv1alpha1.PodStats{
		{
			PodRef: statsv1alpha1.PodReference{Namespace: "default", Name: "app", UID: "uid-a"},
			Containers: []statsv1alpha1.ContainerStats{
				{Name: "app", CPU: &statsv1alpha1.CPUStats{UsageNanoCores: &cpu}, Memory: &statsv1alpha1.MemoryStats{WorkingSetBytes: &memory}},
			},
		},
	}})
	stats, ok := index[containerStatsKey{namespace: "default", podName: "app", podUID: "uid-a", container: "app"}]
	require.True(t, ok)
	assert.Equal(t, int64(250), stats.cpuMillicores)
	assert.Equal(t, int64(1024), stats.memoryBytes)
}

func newTestKubernetesProfileManager(t *testing.T, configs []*KubernetesProfiler) *kubernetesProfileManager {
	t.Helper()
	return newTestKubernetesProfileManagerWithStats(t, configs, &fakeProfileStatsProvider{})
}

func newTestKubernetesProfileManagerWithStats(
	t *testing.T,
	configs []*KubernetesProfiler,
	stats kubeletProfileStatsProvider,
) *kubernetesProfileManager {
	t.Helper()
	manager, err := newKubernetesProfileManager(DefaultInput(), nil, stats, "node-a", configs)
	require.NoError(t, err)
	manager.ctx = context.Background()
	return manager
}

func testProfilePod(uid, podIP, version string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "guancedb",
			Name:      "select-7fdc",
			UID:       types.UID(uid),
			Labels: map[string]string{
				"app":                       "select",
				"app.kubernetes.io/name":    "select",
				"app.kubernetes.io/version": version,
				"pod-template-hash":         "7fdc",
				"team":                      "database",
			},
			OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "select-7fdc"}},
		},
		Spec: corev1.PodSpec{
			NodeName: "node-a",
			Containers: []corev1.Container{
				{
					Name:  "select",
					Ports: []corev1.ContainerPort{{Name: "pprof", ContainerPort: 6060}},
					Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("100"),
					}},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: podIP,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "select", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
			},
		},
	}
}
