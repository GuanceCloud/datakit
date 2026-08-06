// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	containerruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type staticPodMetadataProvider struct {
	pod *corev1.Pod
	err error
}

func (p staticPodMetadataProvider) Get(context.Context, string, string) (*corev1.Pod, error) {
	return p.pod, p.err
}

func TestPodWatcherScopesPodsToLocalNode(t *testing.T) {
	watcher := &podWatcher{nodeName: "node-a"}
	if got := watcher.podListOptions().FieldSelector; got != "spec.nodeName=node-a" {
		t.Fatalf("field selector = %q", got)
	}
	if !watcher.isLocalPod(&corev1.Pod{Spec: corev1.PodSpec{NodeName: "node-a"}}) {
		t.Fatal("local Pod was rejected")
	}
	if watcher.isLocalPod(&corev1.Pod{Spec: corev1.PodSpec{NodeName: "node-b"}}) {
		t.Fatal("remote Pod was accepted")
	}
}

func TestContainerLogUsesCachedPodMetadata(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"datakit/logs": `[{"source":"cached"}]`},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "pod-image"}}},
	}
	item := &containerruntime.Container{
		Name:  "app",
		Image: "runtime-image",
		Labels: map[string]string{
			"io.kubernetes.container.name": "app",
			"io.kubernetes.pod.name":       "example",
			"io.kubernetes.pod.namespace":  "default",
		},
	}
	collector := &containerCollector{podMetadata: staticPodMetadataProvider{pod: pod}}

	info, config := collector.queryContainerLogInfoAndConfig(item)
	if info.image != "pod-image" || config == "" {
		t.Fatalf("cached Pod metadata was not applied: info=%#v config=%q", info, config)
	}
	if info.podLabels == nil || info.podMetadataUnavailable {
		t.Fatal("known empty Pod labels must not look like a cache miss")
	}

	originalPodCacheMissVec := podCacheMissVec
	podCacheMissVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_container_pod_cache_miss_total",
			Help: "Test counter for container Pod metadata cache misses",
		},
		[]string{"usage", "reason"},
	)
	t.Cleanup(func() { podCacheMissVec = originalPodCacheMissVec })

	collector.podMetadata = staticPodMetadataProvider{err: &podCacheMissError{
		reason: podCacheMissNotSynced,
		err:    errors.New("warming"),
	}}
	info, config = collector.queryContainerLogInfoAndConfig(item)
	if info.image != "runtime-image" || config != "" || info.podLabels != nil || !info.podMetadataUnavailable {
		t.Fatalf("cache miss did not preserve runtime metadata: info=%#v config=%q", info, config)
	}
}
