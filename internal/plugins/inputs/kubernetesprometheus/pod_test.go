// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
)

func TestPodUpdatesDoNotRecreateUnchangedTarget(t *testing.T) {
	factory := informers.NewSharedInformerFactory(nil, 0)
	manager := newScrapeManager().(*scrapeManager)
	validator, err := newResourceValidator(nil, "")
	require.NoError(t, err)

	ins := &Instance{
		Role:   string(RolePod),
		Scrape: matchedScrape,
		Target: Target{
			Scheme: "http",
			Port:   "9100",
			Path:   "/metrics",
		},
		Custom: Custom{Tags: map[string]string{}},
	}
	ins.setDefault(&Input{})
	ins.validator = validator

	collector, err := NewPod(factory, []*Instance{ins}, manager, dkio.NewMockedFeeder())
	require.NoError(t, err)
	t.Cleanup(collector.queue.ShutDown)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "default",
			Name:            "example",
			UID:             types.UID("pod-uid"),
			ResourceVersion: "1",
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodRunning,
			HostIP: "192.0.2.1",
			PodIP:  "10.0.0.1",
		},
	}
	require.NoError(t, collector.store.Add(pod))

	key := "default/example"
	collector.queue.Add(key)
	require.True(t, collector.process(context.Background()))

	manager.mu.Lock()
	firstTask := manager.podStore.scrapers[key][0]
	manager.mu.Unlock()
	require.NotNil(t, firstTask)

	updated := pod.DeepCopy()
	updated.ResourceVersion = "2"
	updated.Status.Conditions = []corev1.PodCondition{{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	}}
	require.NoError(t, collector.store.Update(updated))
	collector.queue.Add(key)
	require.True(t, collector.process(context.Background()))

	manager.mu.Lock()
	secondTask := manager.podStore.scrapers[key][0]
	taskCount := len(manager.podStore.scrapers[key])
	manager.mu.Unlock()
	require.Same(t, firstTask, secondTask)
	require.Equal(t, 1, taskCount)

	stopped := updated.DeepCopy()
	stopped.ResourceVersion = "3"
	stopped.Status.Phase = corev1.PodFailed
	require.NoError(t, collector.store.Update(stopped))
	collector.queue.Add(key)
	require.True(t, collector.process(context.Background()))

	manager.mu.Lock()
	taskCount = len(manager.podStore.scrapers[key])
	manager.mu.Unlock()
	require.Zero(t, taskCount)
}
