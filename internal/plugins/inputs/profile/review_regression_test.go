// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func TestReviewProfileURLCredentials(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Location", ":invalid-secret-location")
				writer.WriteHeader(status)
			}))
			defer server.Close()
			endpoint, err := url.Parse(server.URL + "?token=private-token")
			require.NoError(t, err)
			endpoint.User = url.UserPassword("private-user", "private-password")
			profiler := &GoProfiler{URL: endpoint.String(), input: DefaultInput()}
			require.NoError(t, profiler.init())
			_, err = profiler.pullProfileData(context.Background(), "/heap", nil)
			require.Error(t, err)
			for _, secret := range []string{"private-user", "private-password", "private-token", "invalid-secret-location"} {
				assert.NotContains(t, err.Error(), secret)
			}
		})
	}
}

func TestReviewKubernetesRedirectAndFormat(t *testing.T) {
	t.Run("redirect", func(t *testing.T) {
		requests := make(chan struct{}, 1)
		destination := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests <- struct{}{}
			_, err := writer.Write([]byte("internal-response"))
			assert.NoError(t, err)
		}))
		defer destination.Close()
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.Redirect(writer, request, destination.URL, http.StatusFound)
		}))
		defer server.Close()
		manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
		collector, err := manager.newGoCollector(manager.rules[0], server.URL, manager.input)
		require.NoError(t, err)
		_, err = collector.(*goProfileTargetCollector).profiler.pullProfileItem(context.Background(), "goroutine", profileConfigMap["goroutine"], time.Second)
		assert.Error(t, err)
		assert.Empty(t, requests)
	})
	t.Run("invalid profile", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, err := writer.Write([]byte("internal-response"))
			assert.NoError(t, err)
		}))
		defer server.Close()
		manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
		collector, err := manager.newGoCollector(manager.rules[0], server.URL, manager.input)
		require.NoError(t, err)
		_, err = collector.(*goProfileTargetCollector).profiler.pullProfileItem(context.Background(), "goroutine", profileConfigMap["goroutine"], time.Second)
		assert.Error(t, err)
	})
}

func TestReviewKubernetesTagRoundTrip(t *testing.T) {
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	pod.Annotations = map[string]string{"owner": "team,namespace:injected,extra:injected", "service": "custom-service"}
	tags := kubernetesProfileTags(&KubernetesProfiler{PodAnnotationAsTags: map[string]string{"owner": "owner", "service": "service"}}, pod, "select")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(metrics.EventFile, metrics.EventJSONFile)
	require.NoError(t, err)
	require.NoError(t, json.NewEncoder(part).Encode(&metrics.Metadata{TagsProfiler: metrics.JoinTags(tags)}))
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	metadata, _, _, err := DefaultInput().parseProfileMetadata(request, int64(body.Len()))
	require.NoError(t, err)
	assert.Equal(t, "guancedb", metadata["namespace"])
	assert.NotContains(t, metadata, "extra")
	assert.Equal(t, "custom-service", tags["service"])
}

func TestReviewKubernetesNumericPortAmbiguity(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "6060"}})
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: "sidecar"})
	pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
		Name: "sidecar", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
	})
	assert.Len(t, manager.desiredTargets(pod), 1)
	pod.Spec.Containers[1].Ports = pod.Spec.Containers[0].Ports
	assert.Empty(t, manager.desiredTargets(pod), "an ambiguous numeric port must require container selection")
	manager.rules[0].config.Container = "select"
	assert.Len(t, manager.desiredTargets(pod), 1)
}

func TestReviewKubernetesGlobalConcurrency(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{
		{Port: "pprof", Selector: "app.kubernetes.io/version=v1", MaxConcurrency: 2},
		{Port: "pprof", Selector: "app.kubernetes.io/version=v2", MaxConcurrency: 2},
		{Port: "pprof", Selector: "app.kubernetes.io/version=v3", MaxConcurrency: 2},
	})
	collector := &fakeProfileCollector{release: make(chan struct{})}
	defer func() { close(collector.release); manager.workers.Wait() }()
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) { return collector, nil }
	for _, version := range []string{"v1", "v2", "v3"} {
		pod := testProfilePod(version, "10.0.0.8", version)
		pod.Name = version
		pod.Status.PodIP = "10.0.0." + strings.TrimPrefix(version, "v")
		manager.reconcilePod(pod)
	}
	accepted := 0
	for _, target := range manager.targetSnapshot() {
		if manager.dispatch(target, "scheduled", []string{"cpu"}, time.Second, nil, time.Now()) {
			accepted++
		}
	}
	assert.Equal(t, 2, accepted)
}

func TestReviewKubernetesReadyFlap(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof", Cooldown: time.Hour}})
	collector := &fakeProfileCollector{release: make(chan struct{})}
	defer func() { close(collector.release); manager.workers.Wait() }()
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) { return collector, nil }
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	manager.reconcilePod(pod)
	original := manager.targetSnapshot()[0]
	now := time.Now()
	require.True(t, manager.dispatch(original, "resource_threshold", []string{"cpu"}, time.Second, nil, now))
	pod.Status.ContainerStatuses[0].Ready = false
	manager.reconcilePod(pod)
	pod.Status.ContainerStatuses[0].Ready = true
	manager.reconcilePod(pod)
	current := manager.targetSnapshot()[0]
	assert.Same(t, original, current)
	assert.False(t, manager.dispatch(current, "resource_threshold", []string{"cpu"}, time.Second, nil, now.Add(time.Minute)))
}

func TestReviewKubernetesMixedMonitorIntervals(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{
		{Port: "pprof", Selector: "app=unmatched", MonitorInterval: 10 * time.Second, CPUUsageBaseLimit: 80},
		{Port: "pprof", MonitorInterval: 11 * time.Second, TriggerWindow: time.Minute, CPUUsageBaseLimit: 80},
	})
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
	target := manager.targetSnapshot()[0]
	now := target.nextMonitor
	triggered := false
	for elapsed := time.Duration(0); elapsed < 10*time.Minute; {
		for _, due := range manager.dueMonitorTargets(now.Add(elapsed)) {
			if due.evaluateResourceTrigger(resourceUsage{cpuMillicores: 900, cpuLimit: 1000}, now.Add(elapsed)) != nil {
				triggered = true
			}
		}
		elapsed += manager.nextMonitorDelay(now.Add(elapsed))
	}
	assert.True(t, triggered, "a second rule must not disable sustained resource triggers")
}

func TestReviewKubernetesMonitorCancellation(t *testing.T) {
	for _, flushHeaders := range []bool{false, true} {
		t.Run(map[bool]string{false: "headers", true: "body"}[flushHeaders], func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if flushHeaders {
					writer.WriteHeader(http.StatusOK)
					writer.(http.Flusher).Flush()
				}
				close(started)
				select {
				case <-request.Context().Done():
				case <-release:
				}
			}))
			defer server.Close()
			defer close(release)
			stats, err := k8sclient.NewKubeletClient(&rest.Config{}, "http", strings.TrimPrefix(server.URL, "http://"))
			require.NoError(t, err)
			manager := newTestKubernetesProfileManagerWithStats(t, []*KubernetesProfiler{
				{Port: "pprof", CPUUsageBaseLimit: 80},
			}, stats)
			manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			finished := make(chan struct{})
			go func() { manager.runMonitorLoop(ctx); close(finished) }()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("monitor did not request stats")
			}
			cancel()
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("monitor did not stop after cancellation")
			}
		})
	}
}

type retiringProfileCollector struct {
	started  chan struct{}
	canceled chan struct{}
	release  chan struct{}
}

func (collector *retiringProfileCollector) collect(ctx context.Context, _ []string, _ time.Duration, _ map[string]string) error {
	close(collector.started)
	<-ctx.Done()
	close(collector.canceled)
	<-collector.release
	return ctx.Err()
}

func TestReviewKubernetesRetiringEndpointExclusion(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
	collector := &retiringProfileCollector{started: make(chan struct{}), canceled: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(collector.release) }); manager.workers.Wait() }
	defer release()
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) { return collector, nil }
	pod := testProfilePod("uid-a", "10.0.0.8", "v1")
	manager.reconcilePod(pod)
	require.True(t, manager.dispatch(manager.targetSnapshot()[0], "scheduled", []string{"cpu"}, time.Second, nil, time.Now()))
	select {
	case <-collector.started:
	case <-time.After(time.Second):
		t.Fatal("collector did not start")
	}
	manager.deleteObject(pod)
	select {
	case <-collector.canceled:
	case <-time.After(time.Second):
		t.Fatal("deleted target was not canceled")
	}
	pod.UID = "uid-b"
	manager.newCollector = func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error) {
		return &fakeProfileCollector{}, nil
	}
	manager.reconcilePod(pod)
	assert.False(t, manager.dispatch(manager.targetSnapshot()[0], "scheduled", []string{"cpu"}, time.Second, nil, time.Now()),
		"a reused endpoint must wait until the previous worker has completed")
	release()
	assert.True(t, manager.dispatch(manager.targetSnapshot()[0], "scheduled", []string{"cpu"}, time.Second, nil, time.Now()))
}

type deadlineProfileStatsProvider struct {
	remaining   time.Duration
	hasDeadline bool
}

func (provider *deadlineProfileStatsProvider) GetStatsSummaryWithContext(ctx context.Context) (*statsv1alpha1.Summary, error) {
	deadline, exists := ctx.Deadline()
	provider.hasDeadline = exists
	provider.remaining = time.Until(deadline)
	return nil, context.DeadlineExceeded
}

func TestReviewKubernetesStatsDeadline(t *testing.T) {
	for _, interval := range []time.Duration{time.Second, time.Minute} {
		provider := &deadlineProfileStatsProvider{}
		manager := newTestKubernetesProfileManagerWithStats(t, []*KubernetesProfiler{
			{Port: "pprof", CPUUsageBaseLimit: 80, MonitorInterval: interval},
		}, provider)
		manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
		manager.monitor(time.Now())
		assert.True(t, provider.hasDeadline)
		assert.Greater(t, provider.remaining, time.Duration(0))
		assert.LessOrEqual(t, provider.remaining, interval)
		assert.LessOrEqual(t, provider.remaining, 10*time.Second)
	}
}

func TestReviewKubernetesMonitorDelayKeepsPhase(t *testing.T) {
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof", MonitorInterval: 11 * time.Second, CPUUsageBaseLimit: 80}})
	manager.reconcilePod(testProfilePod("uid-a", "10.0.0.8", "v1"))
	target := manager.targetSnapshot()[0]
	now := target.nextMonitor
	require.Len(t, manager.dueMonitorTargets(now), 1)
	assert.Equal(t, 4*time.Second, manager.nextMonitorDelay(now.Add(7*time.Second)))
	require.Len(t, manager.dueMonitorTargets(now.Add(13*time.Second)), 1)
	assert.Equal(t, now.Add(22*time.Second), target.nextMonitor)
}
